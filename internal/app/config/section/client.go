package section

type Client struct {
	Catalog ClientCatalog
}

type ClientCatalog struct {
	GrpcAddr string `envconfig:"CATALOG_SERVICE_GRPC_ADDR" default:"localhost:50051"`
}
