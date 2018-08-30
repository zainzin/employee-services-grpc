package main

import (
	"net"
	"google.golang.org/grpc"
	"./messages"
	"log"
	"context"
	"io"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
)

type server struct{}
var employees = []messages.Employee{
	{Id: 1, BadgeNumber: 1000, FirstName: "Test", LastName: "Tester", VacationAccrualRate: 1.25, VacationAccrued: 21.3},
	{Id:2, BadgeNumber:1001, FirstName:"John", LastName: "Doe", VacationAccrualRate:1.75, VacationAccrued:17.6},
	{Id:3, BadgeNumber:1002, FirstName:"Jane", LastName: "Doe", VacationAccrualRate:1.75, VacationAccrued:16.6},
}
var idCounter int32 = 4
var badgeNumberCounter int32 = 1003

var (
	crt = "certs/cert.pem"
	key = "certs/key.pem"
)

func main() {
	lis, err := net.Listen("tcp", ":9000")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile(crt, key)
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	s := grpc.NewServer(grpc.Creds(creds))
	messages.RegisterEmployeeServicesServer(s, &server{})

	log.Println("Preparing")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (*server) GetByBadgeNumber(ctx context.Context, req *messages.GetByBadgeNumberRequest) (*messages.EmployeeResponse, error) {
	for _, emp := range employees {
		if emp.BadgeNumber == req.BadgeNumber {
			return &messages.EmployeeResponse{
				Employee:&emp,
			}, nil
		}
	}
	return nil, nil
}

func (*server) GetAll(req *messages.GetAllrequest, stream messages.EmployeeServices_GetAllServer) error {
	for _, emp := range employees {
		stream.Send(&messages.EmployeeResponse{
			Employee:&emp,
		})
	}
	return nil
}

func (*server) SaveEmployee(ctx context.Context, req *messages.EmployeeRequest) (*messages.EmployeeResponse, error) {
	emp := req.Employee
	emp.Id = idCounter
	idCounter++
	emp.BadgeNumber = badgeNumberCounter
	badgeNumberCounter++
	employees = append(employees, *emp)

	return &messages.EmployeeResponse{
		Employee:emp,
	}, nil
}

func (*server) SaveAll(stream messages.EmployeeServices_SaveAllServer) error {
	for {
		emp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}

		employees = append(employees, *emp.Employee)
	}
	return nil
}

func (*server) AddPhoto(stream messages.EmployeeServices_AddPhotoServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	log.Println(md)
	if ok {
		log.Println(md["badgenumber"][0])
	}

	var bytes []byte
	for {
		data, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalln(err)
		}
		bytes = append(bytes, data.Data...)

	}
	err := ioutil.WriteFile("./test.png", bytes, 0755)
	if err != nil {
		return err
	}

	log.Println("RECEIVED PHOTO")
	return nil
}
