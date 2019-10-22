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
	Id                 int      `json:"id"`
	UsedMachines       []string `json:"usedMachines"`
	TotalMachinesCount int      `json:"totalMachinesCount"`
}

// PersonRepository is ...
type PersonRepository struct {
	database *sql.DB
}

func UpdatePersonRepository(id int, repository *PersonRepository) int {
	var isWork int
	sqlStatement := `SELECT isWork FROM machine WHERE id=$1;`
	fmt.Println("sqlStatement", sqlStatement)
	row := repository.database.QueryRow(sqlStatement, id)
	err := row.Scan(&isWork)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Zero rows found")
		} else {
			panic(err)
		}
	}
	return isWork
}

// FindAll is ..
func (repository *PersonRepository) FindAll() []*Person {
	var arrUsedMachines []string
	// select machine.id,usedMachines,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id;
	rows, err := repository.database.Query(`select machine.id,usedMachines,totalMachinesCount  from machine  LEFT JOIN loadbalance  on loadbalance_id=loadbalance.id;`)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	people := []*Person{}

	for rows.Next() {
		var (
			id                 int
			usedMachines       string
			totalMachinesCount int
		)
		totalMachinesCount = id
		err = rows.Scan(&id, &usedMachines, &totalMachinesCount)
		if err != nil {
			panic(err)
		}
		if UpdatePersonRepository(id, repository) != 0 {
			arrUsedMachines = append(arrUsedMachines, strconv.Itoa(id))
		}
		//fmt.Println(strings.Split(usedMachines, ","))
		people = append(people, &Person{
			Id:                 id,
			UsedMachines:       arrUsedMachines,
			TotalMachinesCount: id,
		})
	}
	err = rows.Err() // get any error encountered ing iteration
	if err != nil {
		panic(err)
	}
	return people
}

func NewPersonRepository(database *sql.DB) *PersonRepository {
	return &PersonRepository{database: database}
}

type PersonService struct {
	config     *Config
	repository *PersonRepository
}

func (service *PersonService) FindAll() []*Person {
	if service.config.Enabled {
		return service.repository.FindAll()
	}

	return []*Person{}
}

func NewPersonService(config *Config, repository *PersonRepository) *PersonService {
	return &PersonService{config: config, repository: repository}
}

type Server struct {
	config        *Config
	personService *PersonService
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/people", server.findPeople)

	return mux
}

func (server *Server) Run() {
	httpServer := &http.Server{
		Addr:    ":" + server.config.Port,
		Handler: server.Handler(),
	}
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

func (server *Server) findPeople(writer http.ResponseWriter, request *http.Request) {
	people := server.personService.FindAll()
	bytes, _ := json.Marshal(people)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(bytes)
}

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
