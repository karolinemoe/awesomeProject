package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vippsas/vippspoints/keyvault"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/awesomeProject"
	"github.com/sirupsen/logrus"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

var gitSHA string //nolint keyvault

func main() {
	var one int
	var two int64
	var three string
	var four []string
	var five bool
	var six interface{}
	var seven any
	var eight *string

	// logging - til for debugging
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	log.SetOutput(os.Stdout)

	logger := logrus.NewEntry(log).WithFields(logrus.Fields{
		"app":   "awesomeProject",
		"env":   os.Getenv("ENV"),
		"build": gitSHA,
	})

	logger.Info("Hello world")
	logger.Error("Something went wrong, help!")
	logger.Fatal("I'm dying")
	logger.Panicf("Umf..")

	// prometheus/grafana - til for overvåking
	go startPrometheus(logger)
	var err error
	if err != nil {
		awesomeProject.CountError("main")
	}

	// setup api
	var connStr string
	switch os.Getenv("ENV") {
	case "dev":
		connStr = "sqlserver://localhost:1433?user id=SA&password=rV%26%26X5w5c8&database=test"
	default:
		connStr = getFromKeyvault(logger)
	}
	api1, err := awesomeProject.NewPortalAPI(connStr)
	if err != nil {
		// TODO
	}
	router := awesomeProject.NewRouter(api1, awesomeProject.AppAPI{})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", os.Getenv("PORT")),
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Fatal()
	}

	// tips
	// 1. error shadowing
	variable, err := createAnErr()

	newVariable, err := dontCreateErr(variable)
	if err != nil {
		// this will happen
	}

	// 2. Flat struktur til det blir uleselig
	//		- Lag så pakker, jeg er fan av å gruppe etter funksjonalitet.
	//			f.eks. funksjonalitet i en portal
	//					en pakke for innlogin med tilhørende api, domene/service, database
	//					en annen pakke for managing av klippekort med samme filer osv
	//					Ligger alt i en api pakke blir filene mange og store

	// 3. Kommentarer burde ikke være nødvendig IMO, bruk gode funksjonsnavn, argumenter og return statements.
	//		Kan også hjelpe å bruke metoder på structs

	// 4. KISS (Keep it simple stupid)

	// 5. Duplisering kan være bedre om det blir uleselig å lage egne funksjoner

	// 6. TDD, prøv det, gjør koden mye tryggere

	// 7. Azure Devops Pipelines?

	// 8. Ingen map funksjon, må bruke for løkker
	list := []int{1, 2, 3}
	for _, i := range list {
		list = append(list, i+1)
	}

	for i := 0; i <= 10; i++ {
		fmt.Println(i)
	}
}

func startPrometheus(l *logrus.Entry) {
	err := http.ListenAndServe(":9091", promhttp.Handler())
	if err != nil {
		l.Fatal("failed to start prometheus server")
	}
}

func createAnErr() (string, error) {
	return "", errors.New("ERROR")
}

func dontCreateErr(_ string) (string, error) {
	return "", nil
}

func getFromKeyvault(log *logrus.Entry) string {
	azConf := struct {
		AzPodName      string
		AzPodNamespace string
	}{
		AzPodName:      os.Getenv("POD_NAME"),
		AzPodNamespace: os.Getenv("POD_NAMESPACE"),
	}

	vippsnummerClient, err := newKeyVaultClient(azConf, "kv-name")
	if err != nil {
		log.Fatal(err)
	}

	secrets := struct {
		dbConn string `secretName:"mssql-connection-string"`
	}{}
	val := reflect.ValueOf(secrets)
	numFields := val.Elem().NumField()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < numFields; i++ {
		sName := val.Elem().Type().Field(i).Tag.Get("secretName")

		secret, err := vippsnummerClient.Get(ctxTimeout, sName)
		if err != nil {
			log.Fatal(err)
		}

		field := val.Elem().Field(i)
		if !field.CanSet() {
			log.Fatal(err)
		}

		field.SetString(secret)
	}
	return secrets.dbConn
}

type azureClient struct {
	uri    string
	client *keyvault.BaseClient
}

func newKeyVaultClient(cfg azConfig, keyVaultName string) (azureClient, error) {
	kvClient := keyvault.New()

	err := kvClient.AddToUserAgent("golang")
	if err != nil {
		return azureClient{}, errors.Wrap(err, "Unable to add user agent to keyvault client")
	}

	newAuthorizer := func(tenantID, resource string) (*autorest.BearerAuthorizer, error) {
		spt, err := adal.NewServicePrincipalTokenFromManagedIdentity(resource, nil)
		if err != nil {
			return nil, err
		}

		spt.SetSender(adal.CreateSender(newK8sSenderDecorator(cfg.AzPodName, cfg.AzPodNamespace)))

		return autorest.NewBearerAuthorizer(spt), nil
	}
	auth := autorest.NewBearerAuthorizerCallback(kvClient.Sender, newAuthorizer)
	kvClient.Authorizer = auth

	return azureClient{
		uri:    fmt.Sprintf("https://%s.%s", keyVaultName, azure.PublicCloud.KeyVaultDNSSuffix),
		client: &kvClient,
	}, nil
}

type client interface {
	Get(ctx context.Context, name string) (string, error)
}

// Get gets latest version of secret by name. ctx is used for the GetSecret call
func (client azureClient) Get(ctx context.Context, name string) (string, error) {
	version := "" // latest
	uri := client.uri

	s, err := client.client.GetSecret(ctx, uri, name, version)
	if err != nil {
		return "", errors.Wrap(err, "failed to get secret from keyvault")
	}

	if s.Value == nil {
		return "", errors.Errorf("empty secret")
	}

	return *s.Value, nil
}
