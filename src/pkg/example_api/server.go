package example_api

import (
	context "context"
	errors "errors"
	log "log"
	time "time"

	rand "crypto/rand"
	hex "encoding/hex"

	jwt "github.com/pascaldekloe/jwt"
	grpc "google.golang.org/grpc"

	authentication "github.com/dendrite2go/dendrite/src/pkg/authentication"
	axon_utils "github.com/dendrite2go/dendrite/src/pkg/axon_utils"
	axon_server "github.com/dendrite2go/dendrite/src/pkg/grpc/axon_server"
	grpc_config "github.com/dendrite2go/dendrite/src/pkg/grpc/configuration"
	trusted "github.com/dendrite2go/dendrite/src/pkg/trusted"
	utils "github.com/dendrite2go/dendrite/src/pkg/utils"

	grpc_example "github.com/dendrite2go/archetype-go-axon/src/pkg/grpc/example"
)

type GreeterServer struct {
	conn       *grpc.ClientConn
	clientInfo *axon_server.ClientIdentification
}

func (s *GreeterServer) Greet(_ context.Context, greeting *grpc_example.Greeting) (*grpc_example.Acknowledgement, error) {
	message := (*greeting).Message
	log.Printf("Server: Received greeting: %v", message)
	ack := grpc_example.Acknowledgement{
		Message: "Good day to you too!",
	}
	command := grpc_example.GreetCommand{
		AggregateIdentifier: "single_aggregate",
		Message:             greeting,
	}
	if e := axon_utils.SendCommand("GreetCommand", &command, toClientConnection(s)); e != nil {
		return nil, e
	}
	return &ack, nil
}

func (s *GreeterServer) Authorize(_ context.Context, credentials *grpc_example.Credentials) (*grpc_example.AccessToken, error) {
	accessToken := grpc_example.AccessToken{
		Jwt: "",
	}
	if authentication.Authenticate(credentials.Identifier, credentials.Secret) {
		var claims jwt.Claims
		claims.Subject = credentials.Identifier
		claims.Issued = jwt.NewNumericTime(time.Now().Round(time.Second))
		token, e := trusted.CreateJWT(claims)
		if e != nil {
			return nil, e
		}
		accessToken.Jwt = token
	}
	return &accessToken, nil
}

func (s *GreeterServer) ListTrustedKeys(_ *grpc_example.Empty, streamServer grpc_example.GreeterService_ListTrustedKeysServer) error {
	trustedKey := grpc_example.PublicKey{}
	for name, key := range trusted.GetTrustedKeys() {
		trustedKey.Name = name
		trustedKey.PublicKey = key
		log.Printf("Server: Trusted keys streamed reply: %v", trustedKey)
		_ = streamServer.Send(&trustedKey)
		log.Printf("Server: Trusted keys streamed reply sent")
	}
	return nil
}

func (s *GreeterServer) SetPrivateKey(_ context.Context, request *grpc_example.PrivateKey) (*grpc_example.Empty, error) {
	_ = trusted.SetPrivateKey(request.Name, request.PrivateKey)

	var empty = grpc_example.Empty{}
	return &empty, nil
}

func (s *GreeterServer) ChangeTrustedKeys(stream grpc_example.GreeterService_ChangeTrustedKeysServer) error {
	var status = grpc_example.Status{}
	response := grpc_example.TrustedKeyResponse{}
	nonce := make([]byte, 64)
	first := true
	for true {
		request, e := stream.Recv()
		if e != nil {
			log.Printf("Server: Change trusted keys: error receiving request: %v", e)
			return e
		}

		status.Code = 500
		status.Message = "Internal Server Error"

		if first {
			first = false
			status.Code = 200
			status.Message = "OK"
		} else {
			if request.Signature == nil {
				status.Code = 200
				status.Message = "End of stream"
				response.Status = &status
				response.Nonce = nil
				_ = stream.Send(&response)
				return nil
			}
			configRequest := grpc_config.TrustedKeyRequest{}
			if e := utils.ProtoCast(request, &configRequest); e != nil {
				return e
			}
			e = trusted.AddTrustedKey(&configRequest, nonce, toClientConnection(s))
			if e == nil {
				status.Code = 200
				status.Message = "OK"
			} else {
				status.Code = 400
				status.Message = e.Error()
			}
		}

		_, _ = rand.Reader.Read(nonce)
		hexNonce := hex.EncodeToString(nonce)
		log.Printf("Next nonce: %v", hexNonce)

		response.Status = &status
		response.Nonce = nonce
		e = stream.Send(&response)
		if e != nil {
			log.Printf("Server: Change trusted keys: error sending response: %v", e)
			return e
		}
	}
	return errors.New("server: Change trusted keys: unexpected end of stream")
}

func (s *GreeterServer) ChangeCredentials(stream grpc_example.GreeterService_ChangeCredentialsServer) error {
	for true {
		credentials, e := stream.Recv()
		if e != nil {
			log.Printf("Error while receiving credentials: %v", e)
			return e
		}
		if credentials.Signature == nil {
			break
		}
		configCredentials := grpc_config.Credentials{}
		if e := utils.ProtoCast(credentials, &configCredentials); e != nil {
			return e
		}
		_ = authentication.SetCredentials(&configCredentials, toClientConnection(s))
	}
	var empty = grpc_example.Empty{}
	return stream.SendAndClose(&empty)
}

func (s *GreeterServer) SetProperty(_ context.Context, keyValue *grpc_example.KeyValue) (*grpc_example.Empty, error) {
	log.Printf("Server: Set property: %v: %v", keyValue.Key, keyValue.Value)

	command := grpc_example.ChangePropertyCommand{
		Property: keyValue,
	}
	e := axon_utils.SendCommand("ChangePropertyCommand", &command, toClientConnection(s))
	if e != nil {
		log.Printf("Trusted: Error when sending ChangePropertyCommand: %v", e)
	}

	var empty = grpc_example.Empty{}
	return &empty, nil
}

func RegisterWithServer(grpcServer *grpc.Server, clientConnection *axon_utils.ClientConnection) {
	grpc_example.RegisterGreeterServiceServer(grpcServer, &GreeterServer{clientConnection.Connection, clientConnection.ClientInfo})
}

func toClientConnection(s *GreeterServer) *axon_utils.ClientConnection {
	result := axon_utils.ClientConnection{
		Connection: s.conn,
		ClientInfo: s.clientInfo,
	}
	return &result
}
