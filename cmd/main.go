package main

import (
	"context"

	"github.com/IgorRamosBR/g73-techchallenge-payment/configs"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/api"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/controllers"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/core/usecases"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/infra/drivers/dynamodb"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/infra/drivers/http"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/infra/drivers/payment"
	"github.com/IgorRamosBR/g73-techchallenge-payment/internal/infra/gateways"
	"github.com/aws/aws-sdk-go-v2/config"
	awsDynamoDb "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	config := configs.NewConfig()
	appConfig, err := config.ReadConfig()
	if err != nil {
		panic(err)
	}

	paymentHttpClient := http.NewMockHttpClient()
	paymentBrokerConfig := payment.MercadoPagoBrokerConfig{
		HttpClient:      paymentHttpClient,
		BrokerUrl:       appConfig.PaymentBrokerURL,
		NotificationUrl: appConfig.NotificationURL,
		SponsorId:       appConfig.SponsorId,
	}
	paymentBroker := payment.NewMercadoPagoBroker(paymentBrokerConfig)

	dynamodbClient, err := NewDynamoDBClient()
	if err != nil {
		panic(err)
	}
	paymentRepository := gateways.NewPaymentRepositoryGateway(dynamodbClient, appConfig.PaymentTable)

	httpClient := http.NewHttpClient(appConfig.OrderApiTimeout)
	orderClient := gateways.NewOrderClient(httpClient, appConfig.OrderApiUrl)

	paymentUseCaseConfig := usecases.PaymentUseCaseConfig{
		PaymentBroker:     paymentBroker,
		PaymentRepository: paymentRepository,
		OrderClient:       orderClient,
	}
	paymentUseCase := usecases.NewPaymentUseCase(paymentUseCaseConfig)
	paymentController := controllers.NewPaymentController(paymentUseCase)

	api := api.NewApi(paymentController)
	api.Run(":8081")
}

func NewDynamoDBClient() (dynamodb.DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	client := awsDynamoDb.NewFromConfig(cfg)

	return dynamodb.NewDynamoDBClient(client), nil
}
