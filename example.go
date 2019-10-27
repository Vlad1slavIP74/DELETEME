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

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "vlad"
	dbPass := "5"
	dbName := "lab2"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Config is
type Config struct {
	Enabled      bool
	DatabasePath string
	Port         string
}

// NewConfig is
func NewConfig() *Config {
	return &Config{
		Enabled:      true,
		DatabasePath: "./example.db",
		Port:         "8000",
	}
}

// Person is ...
type Person struct {
	Id                 int         `json:"id"`
	UsedMachines       []MachineID `json:"usedMachines"`
	TotalMachinesCount int         `json:"totalMachinesCount"`
}

// MachineID is ...
type MachineID struct {
	MachineID int
}

// PersonRepository is ...
type PersonRepository struct {
	database *sql.DB
}

// Update is ...
func (repository *PersonRepository) Update(isWork string, id string) {
	fmt.Println(isWork)
	fmt.Println(id)
	sqlCall := (`update "machine" set isWork=? where id = ?;`)
	rows, err := repository.database.Prepare(sqlCall)
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}
	rows.Exec(isWork, id)
	// defer repository.database.Close()
}

// FindAll is ..
func (repository *PersonRepository) FindAll() []*Person {
	// select machine.id,usedMachines,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id;
	rows, _ := repository.database.Query(`select loadbalance.id,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id where isWork=1;`)
	defer rows.Close()
	people := []*Person{}
	for rows.Next() {
		machineIDs := []MachineID{}
		machineId := MachineID{}
		var (
			id int
			// usedMachines       string
			totalMachinesCount int
		)
		rows.Scan(&id /*&usedMachines,*/, &totalMachinesCount)
		machineRows, _ := repository.database.Query(`select id from machine where loadbalance_id =` + strconv.Itoa(id) + `;`)
		for machineRows.Next() {
			machineRows.Scan(&id)
			machineId.MachineID = id
			machineIDs = append(machineIDs, machineId)
		}
		people = append(people, &Person{
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
func (service *PersonService) FindAll() []*Person {
	if service.config.Enabled {

		return service.repository.FindAll()
	}

	return []*Person{}
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

	mux.HandleFunc("/people", server.findPeople)
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

func (server *Server) findPeople(writer http.ResponseWriter, request *http.Request) {
	people := server.personService.FindAll()
	bytes, _ := json.Marshal(people)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(bytes)
}
func (server *Server) updateMachine(writer http.ResponseWriter, request *http.Request) {
	fmt.Println(request.FormValue("id"))
	isWork := request.FormValue("isWork")
	id := request.FormValue("id")
	server.personService.Update(isWork, id)
	writer.WriteHeader(http.StatusOK)
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
	container := BuildContainer()

	err := container.Invoke(func(server *Server) {
		server.Run()
	})

	if err != nil {
		panic(err)
	}
}

// The manual way
//
// func main() {
// 	config := NewConfig()
//
// 	db, err := ConnectDatabase(config)
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	personRepository := NewPersonRepository(db)
//
// 	personService := NewPersonService(config, personRepository)
//
// 	server := NewServer(config, personService)
//
// 	server.Run()
// }
