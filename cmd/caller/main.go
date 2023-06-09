package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/resource"
)

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                getEnv("SOLACE_HOST", "tcp://localhost:55554"),
		config.ServicePropertyVPNName:                    getEnv("SOLACE_VPN", "default"),
		config.AuthenticationPropertySchemeBasicPassword: getEnv("SOLACE_PASSWORD", "default"),
		config.AuthenticationPropertySchemeBasicUserName: getEnv("SOLACE_USERNAME", "default"),
	}

	messagingService, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(brokerConfig).
		WithTransportSecurityStrategy(config.NewTransportSecurityStrategy().WithoutCertificateValidation()).
		Build()

	if err != nil {
		panic(err)
	}

	// Connect to the messaging serice
	if err := messagingService.Connect(); err != nil {
		panic(err)
	}

	defer messagingService.Disconnect()

	fmt.Println("Connected to the broker? ", messagingService.IsConnected())

	publisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		panic(err)
	}

	defer publisher.Terminate(10 * time.Second)

	if err := publisher.Start(); err != nil {
		panic(err)
	}

	msg, _ := messagingService.MessageBuilder().BuildWithStringPayload("call request")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sig:
				fmt.Println("SIGTERM received. Exiting...")
				os.Exit(0)
			default:
				a := getRandomPhoneNumber()
				b := getRandomPhoneNumber()

				fmt.Printf("calling... %s -> %s\n", a, b)

				if err := publisher.Publish(msg, resource.TopicOf("af4092/Call/"+a+"/"+b)); err != nil {
					panic(err)
				}

				time.Sleep(1 * time.Second)
			}
		}
	}()

	for {
	}
}

func getRandomPhoneNumber() string {
	return fmt.Sprintf("010%04d%04d", time.Now().UnixNano()%10000, time.Now().UnixNano()%10000)
}
