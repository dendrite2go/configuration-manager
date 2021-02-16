package main

import (
	fmt "fmt"
	log "log"

	uuid "github.com/google/uuid"
	grpc "google.golang.org/grpc"

	authentication "github.com/dendrite2go/dendrite/src/pkg/authentication"
	axon_utils "github.com/dendrite2go/dendrite/src/pkg/axon_utils"
	configuration_api "github.com/dendrite2go/dendrite/src/pkg/configuration_api"
	configuration_query "github.com/dendrite2go/dendrite/src/pkg/configuration_query"
	axon_server "github.com/dendrite2go/dendrite/src/pkg/grpc/axon_server"
	trusted "github.com/dendrite2go/dendrite/src/pkg/trusted"
	utils "github.com/dendrite2go/dendrite/src/pkg/utils"

	cache_utils "github.com/dendrite2go/archetype-go-axon/src/pkg/cache_utils"
	example_api "github.com/dendrite2go/archetype-go-axon/src/pkg/example_api"
	example_command "github.com/dendrite2go/archetype-go-axon/src/pkg/example_command"
	example_query "github.com/dendrite2go/archetype-go-axon/src/pkg/example_query"
	example_trusted "github.com/dendrite2go/archetype-go-axon/src/pkg/trusted"
)

func main() {
	log.Printf("\n\n\n")
	log.Printf("Start Go Client")

	example_trusted.Init()
	authentication.Init()
	for k, v := range trusted.GetTrustedKeys() {
		log.Printf("Trusted key: %v: %v", k, v)
	}

	host := "proxy"
	port := 8124
	clientConnection, streamClient := axon_utils.WaitForServer(host, port, "API")
	defer utils.ReportError("Close clientConnection", clientConnection.Connection.Close)
	log.Printf("Main connection: %v: %v", clientConnection, streamClient)

	// Send a heartbeat
	heartbeat := axon_server.Heartbeat{}
	heartbeatRequest := axon_server.PlatformInboundInstruction_Heartbeat{
		Heartbeat: &heartbeat,
	}
	id := uuid.New()
	instruction := axon_server.PlatformInboundInstruction{
		Request:       &heartbeatRequest,
		InstructionId: id.String(),
	}
	if e := (*streamClient).Send(&instruction); e != nil {
		panic(fmt.Sprintf("Error sending clientInfo %v", e))
	}

	// Initialize cache
	cache_utils.InitializeCache()

	// Handle commands
	commandHandlerConn := example_command.HandleCommands(host, port)
	defer utils.ReportError("Close commandHandlerConn", commandHandlerConn.Connection.Close)

	// Process Events
	eventProcessorConn := example_query.ProcessEvents(host, port)
	defer utils.ReportError("Close eventProcessorConn", eventProcessorConn.Connection.Close)

	configurationEventProcessorConn := configuration_query.ProcessEvents(host, port)
	defer utils.ReportError("Close configurationEventProcessorConn", configurationEventProcessorConn.Connection.Close)

	// Handle queries
	queryHandlerConn := example_query.HandleQueries(host, port)
	defer utils.ReportError("Close queryHandlerConn", queryHandlerConn.Connection.Close)

	// Listen to incoming gRPC requests
	_ = axon_utils.Serve(clientConnection, registerWithServer)
}

func registerWithServer(server *grpc.Server, conn *axon_utils.ClientConnection) {
	example_api.RegisterWithServer(server, conn)
	configuration_api.RegisterWithServer(server, conn)
}
