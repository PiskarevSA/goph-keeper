package gophkeeper

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	gophkeeperv1 "github.com/PiskarevSA/goph-keeper/gen/gophkeeper/v1"
	"github.com/PiskarevSA/goph-keeper/internal/auth"
	"github.com/PiskarevSA/goph-keeper/internal/domain"
	"github.com/PiskarevSA/goph-keeper/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
	gophkeeperv1.UnimplementedUserServiceServer
	users *service.UserService
	jwt   *auth.JWTManager
}

func NewUserServer(users *service.UserService, jwt *auth.JWTManager,
) *UserServer {
	return &UserServer{
		users: users,
		jwt:   jwt,
	}
}

func (s *UserServer) Register(
	ctx context.Context, req *gophkeeperv1.RegisterRequest,
) (*gophkeeperv1.RegisterResponse, error) {
	userID, err := s.users.Register(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := s.jwt.Generate(userID)
	if err != nil {
		return nil, err
	}
	return &gophkeeperv1.RegisterResponse{Token: token}, nil
}

func (s *UserServer) Login(
	ctx context.Context, req *gophkeeperv1.LoginRequest,
) (*gophkeeperv1.LoginResponse, error) {
	user, err := s.users.Authenticate(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, err
	}
	return &gophkeeperv1.LoginResponse{Token: token}, nil
}

type SecretsServer struct {
	gophkeeperv1.UnimplementedSecretsServiceServer
	secrets *service.SecretsService
}

func NewSecretsServer(secrets *service.SecretsService) *SecretsServer {
	return &SecretsServer{secrets: secrets}
}

func (s *SecretsServer) InfoList(
	ctx context.Context, req *gophkeeperv1.InfoListRequest,
) (*gophkeeperv1.InfoListResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	list, err := s.secrets.List(userID)
	if err != nil {
		return nil, err
	}
	resp := &gophkeeperv1.InfoListResponse{}
	for _, sec := range list {
		resp.SecretsInfo = append(resp.SecretsInfo, &gophkeeperv1.SecretInfo{
			Uuid:     sec.UUID,
			Name:     sec.Name,
			Kind:     domainToPbInfoKind(sec.Kind),
			Created:  timestamppb.New(sec.Created),
			Modified: timestamppb.New(sec.Modified),
		})
	}
	return resp, nil
}

func (s *SecretsServer) Create(
	ctx context.Context, req *gophkeeperv1.CreateRequest,
) (*gophkeeperv1.CreateResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	sec := domain.Secret{
		Name:    req.GetName(),
		Kind:    deriveKindFromPbDetails(req.GetDetails()),
		Details: pbDetailsToDomain(req.GetDetails()),
	}

	created, err := s.secrets.Create(userID, sec)
	if err != nil {
		return nil, err
	}

	return &gophkeeperv1.CreateResponse{
		Info: &gophkeeperv1.SecretInfo{
			Uuid:     created.UUID,
			Name:     created.Name,
			Kind:     domainToPbInfoKind(created.Kind),
			Created:  timestamppb.New(created.Created),
			Modified: timestamppb.New(created.Modified),
		},
	}, nil
}

func (s *SecretsServer) ReadDetails(
	ctx context.Context, req *gophkeeperv1.ReadDetailsRequest,
) (*gophkeeperv1.ReadDetailsResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	sec, err := s.secrets.Get(userID, req.GetInfo().GetUuid())
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	err = checkFresh(req.GetInfo().GetModified().AsTime(), sec.Modified)
	if err != nil {
		return nil, err
	}

	return &gophkeeperv1.ReadDetailsResponse{
		Details: domainDetailsToPb(sec.Details),
	}, nil
}

func (s *SecretsServer) Update(
	ctx context.Context, req *gophkeeperv1.UpdateRequest,
) (*gophkeeperv1.UpdateResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	sec, err := s.secrets.Get(userID, req.GetInfo().GetUuid())
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	err = checkFresh(req.GetInfo().GetModified().AsTime(), sec.Modified)
	if err != nil {
		return nil, err
	}

	sec.Name = req.GetInfo().GetName()
	sec.Kind = pbInfoKindToDomain(req.GetInfo().GetKind())
	sec.Details = pbDetailsToDomain(req.GetDetails())

	updated, err := s.secrets.Update(userID, *sec)
	if err != nil {
		return nil, err
	}

	return &gophkeeperv1.UpdateResponse{Modified: timestamppb.New(updated)}, nil
}

func (s *SecretsServer) Delete(
	ctx context.Context, req *gophkeeperv1.DeleteRequest,
) (*gophkeeperv1.DeleteResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	sec, err := s.secrets.Get(userID, req.GetInfo().GetUuid())
	if err != nil {
		return nil, err
	}
	if sec == nil {
		return nil, status.Error(codes.NotFound, "secret not found")
	}

	err = checkFresh(req.GetInfo().GetModified().AsTime(), sec.Modified)
	if err != nil {
		return nil, err
	}

	if err := s.secrets.Delete(userID, sec.UUID); err != nil {
		return nil, err
	}

	return &gophkeeperv1.DeleteResponse{}, nil
}

func (s *SecretsServer) UploadRaw(
	stream gophkeeperv1.SecretsService_UploadRawServer,
) error {
	userID, ok := auth.UserIDFromContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	var uuid string
	var expectedModified time.Time
	var tmpFilePath string
	var outFile *os.File

	defer func() {
		if outFile != nil {
			outFile.Close()
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch v := req.Data.(type) {
		case *gophkeeperv1.UploadRawRequest_Begin_:
			uuid = v.Begin.Uuid
			expectedModified = v.Begin.ExpectedModified.AsTime()

			sec, err := s.secrets.Get(userID, uuid)
			if err != nil {
				return err
			}
			if sec == nil {
				return status.Error(codes.NotFound, "secret not found")
			}
			if !sec.Modified.Equal(expectedModified) {
				return status.Error(codes.FailedPrecondition, "secret info is outdated")
			}

			tmpDir := "./data/raw"
			if err = os.MkdirAll(tmpDir, 0o755); err != nil {
				return err
			}
			tmpFilePath = filepath.Join(tmpDir, uuid)
			outFile, err = os.Create(tmpFilePath)
			if err != nil {
				return err
			}

		case *gophkeeperv1.UploadRawRequest_Chunk:
			if outFile == nil {
				return status.Error(codes.FailedPrecondition, "upload not initialized")
			}
			if _, err := outFile.Write(v.Chunk.Content); err != nil {
				return err
			}
		}
	}

	if outFile != nil {
		outFile.Close()
	}

	modified, err := s.secrets.SaveRawFile(userID, uuid, tmpFilePath)
	if err != nil {
		return err
	}

	return stream.SendAndClose(&gophkeeperv1.UploadRawResponse{
		Modified: timestamppb.New(modified),
	})
}

func (s *SecretsServer) DownloadRaw(
	req *gophkeeperv1.DownloadRawRequest,
	stream gophkeeperv1.SecretsService_DownloadRawServer,
) error {
	userID, ok := auth.UserIDFromContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	sec, err := s.secrets.Get(userID, req.GetUuid())
	if err != nil {
		return err
	}
	if sec == nil || sec.Details.Raw == nil {
		return status.Error(codes.NotFound, "raw secret not found")
	}

	f, err := os.Open(sec.Details.Raw.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	// отправляем метаданные
	if err := stream.Send(&gophkeeperv1.DownloadRawResponse{
		Data: &gophkeeperv1.DownloadRawResponse_Begin_{
			Begin: &gophkeeperv1.DownloadRawResponse_Begin{
				Filename: sec.Details.Raw.Filename,
				Size:     sec.Details.Raw.Size,
				Modified: timestamppb.New(sec.Modified),
			},
		},
	}); err != nil {
		return err
	}

	buf := make([]byte, 64*1024) // 64KB chunk
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&gophkeeperv1.DownloadRawResponse{
			Data: &gophkeeperv1.DownloadRawResponse_Chunk{
				Chunk: &gophkeeperv1.FileChunk{Content: buf[:n]},
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

// checkFresh проверяет, что клиентская версия секрета актуальна
func checkFresh(client, actual time.Time) error {
	if !client.Equal(actual) {
		return status.Error(codes.FailedPrecondition, "secret info is outdated")
	}
	return nil
}
