package util

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/model"
)

func FindDesiredStateServer(servers []config.DesiredStateServer, serverName string) (*config.DesiredStateServer, error) {
	for _, server := range servers {
		if server.Name == serverName {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("desired state server %s does not exist", serverName)
}

func ForeignServerFromDesiredStateServer(server config.DesiredStateServer) model.ForeignServer {
	return model.ForeignServer{
		Name:    server.Name,
		Host:    server.Host,
		Port:    server.Port,
		DB:      server.DB,
		Wrapper: server.Wrapper,
	}
}

func ForeignServersFromDesiredStateServers(servers []config.DesiredStateServer) []model.ForeignServer {
	fservers := make([]model.ForeignServer, 0)
	for _, server := range servers {
		fservers = append(fservers, ForeignServerFromDesiredStateServer(server))
	}
	return fservers
}

func UserMapFromDesiredStateUserMap(serverName string, usermap config.DesiredStateUserMap) model.UserMap {
	return model.UserMap{
		ServerName:     serverName,
		LocalUser:      usermap.LocalUser,
		RemoteUser:     usermap.RemoteUser,
		RemotePassword: usermap.RemotePassword,
	}
}

func UserMapsFromDesiredStateUserMaps(serverName string, usermaps []config.DesiredStateUserMap) []model.UserMap {
	umaps := make([]model.UserMap, 0)
	for _, usermap := range usermaps {
		umaps = append(umaps, UserMapFromDesiredStateUserMap(serverName, usermap))
	}
	return umaps
}

func SchemaFromDesiredStateSchema(dsSchema config.DesiredStateSchema) model.Schema {
	return model.Schema{
		LocalSchema:    dsSchema.LocalSchema,
		RemoteSchema:   dsSchema.RemoteSchema,
		ImportENUMs:    dsSchema.ImportENUMs,
		ENUMConnection: dsSchema.ENUMConnection,
	}
}

func SchemasFromDesiredStateSchemas(dsSchemas []config.DesiredStateSchema) []model.Schema {
	schemas := make([]model.Schema, 0)
	for _, dsSchema := range dsSchemas {
		schemas = append(schemas, SchemaFromDesiredStateSchema(dsSchema))
	}
	return schemas
}
