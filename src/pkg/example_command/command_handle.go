package example_command

import (
	context "context"
	log "log"

	authentication "github.com/dendrite2go/dendrite/src/pkg/authentication"
	axon_utils "github.com/dendrite2go/dendrite/src/pkg/axon_utils"
	configuration_command "github.com/dendrite2go/dendrite/src/pkg/configuration_command"
	axon_server "github.com/dendrite2go/dendrite/src/pkg/grpc/axon_server"
	trusted "github.com/dendrite2go/dendrite/src/pkg/trusted"
)

func HandleCommands(host string, port int) *axon_utils.ClientConnection {
	clientConnection, _ := axon_utils.WaitForServer(host, port, "Command Handler")
	conn := clientConnection.Connection
	clientInfo := clientConnection.ClientInfo

	log.Printf("Command handler: Connection: %v", conn)
	client := axon_server.NewCommandServiceClient(conn)
	log.Printf("Command handler: Client: %v", client)

	stream, e := client.OpenStream(context.Background())
	log.Printf("Command handler: Stream: %v: %v", stream, e)

	axon_utils.SubscribeCommand("RegisterTrustedKeyCommand", stream, clientInfo)
	axon_utils.SubscribeCommand("RegisterKeyManagerCommand", stream, clientInfo)
	axon_utils.SubscribeCommand("RegisterCredentialsCommand", stream, clientInfo)
	axon_utils.SubscribeCommand("ChangePropertyCommand", stream, clientInfo)

	go axon_utils.CommandWorker(stream, clientConnection, commandDispatch)

	return clientConnection
}

func commandDispatch(command *axon_server.Command, stream axon_server.CommandService_OpenStreamClient, clientConnection *axon_utils.ClientConnection) (*axon_utils.Error, error) {
	commandName := command.Name
	if commandName == "RegisterTrustedKeyCommand" {
		return trusted.HandleRegisterTrustedKeyCommand(command, clientConnection)
	} else if commandName == "RegisterKeyManagerCommand" {
		return trusted.HandleRegisterKeyManagerCommand(command, clientConnection)
	} else if commandName == "RegisterCredentialsCommand" {
		return authentication.HandleRegisterCredentialsCommand(command, clientConnection)
	} else if commandName == "ChangePropertyCommand" {
		return configuration_command.HandleChangePropertyCommand(command, clientConnection)
	} else {
		log.Printf("Received unknown command: %v", commandName)
	}
	return nil, nil
}
