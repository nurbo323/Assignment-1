package main

import (
	"context"
	"log"
	"payment-service/internal/app"
	repopg "payment-service/internal/repository/postgres"
	grpctransport "payment-service/internal/transport/grpc"
	httptransport "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := app.LoadConfig()

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	repo := repopg.NewPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	handler := httptransport.NewPaymentHandler(uc)

	go func() {
		if err := grpctransport.RunGRPCServer(uc); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	router := gin.Default()
	handler.RegisterRoutes(router)

	log.Printf("payment-service listening on :%s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
