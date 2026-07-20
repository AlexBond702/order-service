package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofrs/uuid/v5"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/AlexBond702/order-service/internal/catalog/gen/v1"
)

type CatalogClient struct {
	conn   *grpc.ClientConn
	client pb.CatalogServiceClient
}

func NewCatalogClient(addr string) (*CatalogClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
		grpc.WithChainStreamInterceptor(grpc_prometheus.StreamClientInterceptor))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to catalog service: %w", err)
	}
	client := pb.NewCatalogServiceClient(conn)
	log.Printf("Connected to Catalog Service at %s", addr)
	return &CatalogClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *CatalogClient) Close() error {
	if c.conn != nil {
		log.Print("Closing connection to Catalog Service")
		return c.conn.Close()
	}
	return nil
}

type ProductInfo struct {
	ID           string
	Guid         string
	Name         string
	Description  string
	Price        float64
	CategoryGuid string
}

func (c *CatalogClient) GetProducts(ctx context.Context, productsIds []uuid.UUID) (map[string]*ProductInfo, error) {
	guids := make([]string, len(productsIds))
	for i, id := range productsIds {
		guids[i] = id.String()
	}
	req := &pb.GetProductsRequest{
		Guids: guids,
	}
	resp, err := c.client.GetProducts(ctx, req)
	if err != nil {
		return nil, c.handleError(err)
	}

	products := make(map[string]*ProductInfo)
	for _, product := range resp.Products {
		products[product.Guid] = &ProductInfo{
			ID:           strconv.FormatInt(product.Id, 10),
			Guid:         product.Guid,
			Name:         product.Name,
			Description:  *product.Description,
			Price:        product.Price,
			CategoryGuid: product.CategoryGuid,
		}
	}
	return products, nil
}

func (c *CatalogClient) CheckProductExist(ctx context.Context, productId uuid.UUID) (float64, error) {
	productStr := productId.String()

	req := &pb.CheckProductExistsRequest{
		Guid: productStr,
	}

	resp, err := c.client.CheckProductExists(ctx, req)
	if err != nil {
		return 0, c.handleError(err)
	}
	return resp.Product.Price, nil
}

func (c *CatalogClient) handleError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("catalog service error: %w", err)
	}
	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("product not found in catalog")
	case codes.Unavailable:
		return fmt.Errorf("catalog service unavailable")
	case codes.InvalidArgument:
		return fmt.Errorf("invalid request to catalog service: %s", st.Message())
	default:
		return fmt.Errorf("catalog service error [%s]: %s", st.Code(), st.Message())
	}
}
