package service

import (
	"context"
	"database/sql"
	"log"

	"AliveVirtualGift_SessionService/src/proto"
	"AliveVirtualGift_SessionService/src/utils"
)

//serviceServer ...
type serviceServer struct {
	db *sql.DB
}

//NewSessionServiceServer ...
func NewSessionServiceServer(db *sql.DB) proto.SessionServiceServer {
	return &serviceServer{db: db}
}

func (s *serviceServer) GetAccountIDFromToken(ctx context.Context, request *proto.TokenString) (*proto.AccountID, error) {

	claims := utils.ExtractClaims(request.GetToken())
	accInfo := claims["account_info"]

	accInfoMap, ok := accInfo.(map[string]interface{})
	if !ok {
		log.Print("Payload Conversion failed")
	}
	accInfoMap["id"] = uint64(accInfoMap["id"].(float64))

	return &proto.AccountID{
		Id: accInfoMap["id"].(uint64),
	}, nil
}

func (s *serviceServer) GetAccountTypeFromToken(ctx context.Context, request *proto.TokenString) (*proto.AccountType, error) {

	claims := utils.ExtractClaims(request.GetToken())
	accInfo := claims["account_info"]

	accInfoMap, ok := accInfo.(map[string]interface{})
	if !ok {
		log.Print("Payload Conversion failed")
	}
	accInfoMap["type"] = int32(accInfoMap["type"].(float64))

	return &proto.AccountType{
		Type: proto.Type(accInfoMap["type"].(int32)),
	}, nil
}

func (s *serviceServer) CreateToken(ctx context.Context, request *proto.AccountInfo) (*proto.TokenString, error) {

	td, err := utils.GenerateToken(request)
	if err != nil {
		return nil, err
	}

	err = utils.CreateAuth(request.GetId(), td)
	if err != nil {
		return nil, err
	}

	return &proto.TokenString{
		Token: td.AccessToken,
	}, nil
}

func (s *serviceServer) RefreshToken(ctx context.Context, request *proto.TokenString) (*proto.TokenString, error) {
	return nil, nil
}

func (s *serviceServer) DeleteToken(ctx context.Context, request *proto.TokenString) (*proto.Status, error) {
	return nil, nil
}

func (s *serviceServer) CheckToken(ctx context.Context, request *proto.TokenString) (*proto.Status, error) {
	return nil, nil
}
