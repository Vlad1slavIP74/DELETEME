package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/dig"
)

// Config is
type Config struct {
	Enabled      bool
	DatabasePath string
	Port         string
}

// NewConfig is constructor
// DatabasePath - path to our db
func NewConfig() *Config {
	return &Config{
		Enabled:      true,
		DatabasePath: "./example.db",
		Port:         "8000",
	}
}

// LoadBalancer is structure which will be returned as a json
type LoadBalancer struct {
	Id                 int   `json:"id"`
	UsedMachines       []int `json:"usedMachines"`
	TotalMachinesCount int   `json:"totalMachinesCount"`
}

// PersonRepository is our db
type PersonRepository struct {
	database *sql.DB
}

// Update is our machine
func (repository *PersonRepository) Update(isWork string, id string) {
	sqlCall := (`update "machine" set isWork=? where id = ?;`)
	rows, err := repository.database.Prepare(sqlCall)
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}
	rows.Exec(isWork, id)
}

// FindAll is function that shows all load
func (repository *PersonRepository) FindAll() []*LoadBalancer {
	// select machine.id,usedMachines,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id;
	rows, _ := repository.database.Query(`select loadbalance.id,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id;`)
	defer rows.Close()
	people := []*LoadBalancer{}
	for rows.Next() {
		machineIDs := []int{}
		// machineId := MachineID{}
		var (
			id int
			// usedMachines       string
			totalMachinesCount int
		)
		rows.Scan(&id /*&usedMachines,*/, &totalMachinesCount)
		machineRows, _ := repository.database.Query(`select id from machine where loadbalance_id =` + strconv.Itoa(id) + ` AND isWork = 1;`)
		if machineRows != nil {
			for machineRows.Next() {
				machineRows.Scan(&id)
				// machineId.MachineID = id
				machineIDs = append(machineIDs, id)
			}
		}
		people = append(people, &LoadBalancer{
			Id:                 id,
			UsedMachines:       machineIDs, //[]string{usedMachines},
			TotalMachinesCount: totalMachinesCount,
		})
	}

	return people
}

// NewPersonRepository is ...
func NewPersonRepository(database *sql.DB) *PersonRepository {
	return &PersonRepository{database: database}
}

type PersonService struct {
	config     *Config
	repository *PersonRepository
	updates    *PersonRepository
}

// FindAll is....
func (service *PersonService) FindAll() []*LoadBalancer {
	if service.config.Enabled {

		return service.repository.FindAll()
	}

	return []*LoadBalancer{}
}

// Update is...
func (service *PersonService) Update(isWork string, id string) {
	if service.config.Enabled {
		service.repository.Update(isWork, id) //.Update(isWork, id)
	}
}

// NewPersonService is
func NewPersonService(config *Config, repository *PersonRepository) *PersonService {
	return &PersonService{config: config, repository: repository}
}

// Server is ...
type Server struct {
	config        *Config
	personService *PersonService
	updates       *PersonService
}

// Handler is ...
func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/list", server.findBalancer)
	mux.HandleFunc("/update", server.updateMachine)
	return mux
}

func (server *Server) Run() {
	httpServer := &http.Server{
		Addr:    ":" + server.config.Port,
		Handler: server.Handler(),
	}

	httpServer.ListenAndServe()
}

func (server *Server) findBalancer(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		people := server.personService.FindAll()
		bytes, _ := json.Marshal(people)

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(bytes)
	} else {
		fmt.Println("Method should be GET")
	}

}
func (server *Server) updateMachine(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut {
		isWork := request.FormValue("isWork")
		id := request.FormValue("id")
		fmt.Println("ID")
		fmt.Println(id)
		fmt.Println("ISWORK")
		fmt.Println(isWork)
		server.personService.Update(isWork, id)
		writer.WriteHeader(http.StatusOK)
	} else {
		fmt.Println("Method should be PUT")
	}

}

// NewServer is...
func NewServer(config *Config, personService *PersonService) *Server {
	return &Server{
		config:        config,
		personService: personService,
	}
}

func ConnectDatabase(config *Config) (*sql.DB, error) {
	return sql.Open("sqlite3", config.DatabasePath)
}

func BuildContainer() *dig.Container {
	container := dig.New()

	container.Provide(NewConfig)
	container.Provide(ConnectDatabase)
	container.Provide(NewPersonRepository)
	container.Provide(NewPersonService)
	container.Provide(NewServer)

	return container
}

func main() {
	fmt.Println("Server has been started at the port 8000...")
	container := BuildContainer()

	err := container.Invoke(func(server *Server) {
		server.Run()
	})
	if err != nil {
		panic(err)
	}
}
