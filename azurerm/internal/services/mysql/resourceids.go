package mysql

//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Key -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1/keys/key1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Server -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/servers/server1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServer -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServerKey -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1/keys/key1
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=FlexibleServerFirewallRule -id=/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.DBforMySQL/flexibleServers/flexibleServer1/firewallRules/firewallRule1
