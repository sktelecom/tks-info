package app_serve_app

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-info/pkg/app_serve_app/model"
	pb "github.com/openinfradev/tks-proto/tks_pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Accessor is an accessor to postgreSQL to query data.
type AsaAccessor struct {
	db *gorm.DB
}

// New returns new accessor's ptr.
func New(db *gorm.DB) *AsaAccessor {
	return &AsaAccessor{
		db: db,
	}
}

// Create creates a new appServeApp in database.
func (x *AsaAccessor) Create(contractId string, app *pb.AppServeApp, task *pb.AppServeAppTask) (uuid.UUID, uuid.UUID, error) {
	asaModel := model.AppServeApp{
		Name:            app.GetName(),
		ContractId:      contractId,
		TaskType:        app.GetTaskType(),
		TargetClusterId: app.GetTargetClusterId(),
	}

	res := x.db.Create(&asaModel)
	if res.Error != nil {
		return uuid.Nil, uuid.Nil, res.Error
	}

	asaTaskModel := model.AppServeAppTask{
		Version:       task.GetVersion(),
		Status:        task.GetStatus(),
		ArtifactUrl:   task.GetArtifactUrl(),
		ImageUrl:      task.GetImageUrl(),
		Profile:       task.GetProfile(),
		AppServeAppId: asaModel.ID,
	}

	res = x.db.Create(&asaTaskModel)
	if res.Error != nil {
		return uuid.Nil, uuid.Nil, res.Error
	}

	return asaModel.ID, asaTaskModel.ID, nil
}

// Update creates new appServeApp Task for existing appServeApp.
func (x *AsaAccessor) Update(appServeAppId uuid.UUID, task *pb.AppServeAppTask) (uuid.UUID, error) {
	asaTaskModel := model.AppServeAppTask{
		Version:       task.GetVersion(),
		Status:        task.GetStatus(),
		ArtifactUrl:   task.GetArtifactUrl(),
		ImageUrl:      task.GetImageUrl(),
		Profile:       task.GetProfile(),
		AppServeAppId: appServeAppId,
	}

	res := x.db.Create(&asaTaskModel)
	if res.Error != nil {
		return uuid.Nil, res.Error
	}

	return asaTaskModel.ID, nil
}

func (x *AsaAccessor) GetAppServeApps(contractId string) ([]*pb.AppServeApp, error) {
	var appServeApps []model.AppServeApp
	pbAppServeApps := []*pb.AppServeApp{}

	res := x.db.Find(&appServeApps, "contract_id = ?", contractId)
	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding appServeApps with contractID: %s", contractId)
	}

	// If no record is found, just return empty array.
	if res.RowsAffected == 0 {
		return pbAppServeApps, nil
	}

	for _, asa := range appServeApps {
		pbAppServeApps = append(pbAppServeApps, ConvertToPbAppServeApp(asa))
	}
	return pbAppServeApps, nil
}

func (x *AsaAccessor) GetAppServeApp(id uuid.UUID) (*pb.AppServeAppCombined, error) {
	var appServeApp model.AppServeApp
	var appServeAppTasks []model.AppServeAppTask
	pbAppServeAppCombined := &pb.AppServeAppCombined{}

	res := x.db.First(&appServeApp, "id = ?", id)
	if res.RowsAffected == 0 || res.Error != nil {
		return nil, fmt.Errorf("Could not find AppServeApp with ID: %s", id)
	}
	pbAppServeAppCombined.AppServeApp = ConvertToPbAppServeApp(appServeApp)

	res = x.db.Order("created_at asc").Find(&appServeAppTasks, "app_serve_app_id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("Error while finding appServeAppTasks with appServeApp ID %s. Err: %s", id, res.Error)
	}

	for _, task := range appServeAppTasks {
		pbAppServeAppCombined.Tasks = append(pbAppServeAppCombined.Tasks, ConvertToPbAppServeAppTask(task))
	}

	return pbAppServeAppCombined, nil
}

func (x *AsaAccessor) UpdateStatus(taskId uuid.UUID, status string, output string) error {
	res := x.db.Model(&model.AppServeAppTask{}).Where("ID = ?", taskId).Updates(model.AppServeAppTask{Status: status, Output: output})

	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: nothing updated in AppServeAppTask with ID %s", taskId)
	}

	return nil
}

func (x *AsaAccessor) UpdateEndpoint(id uuid.UUID, taskId uuid.UUID, endpoint string, helmRevision int32) error {
	// Update Endpoint
	res := x.db.Model(&model.AppServeApp{}).Where("ID = ?", id).Update("EndpointUrl", endpoint)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeApp with id %s", id)
	}

	// Update helm revision
	res = x.db.Model(&model.AppServeAppTask{}).Where("ID = ?", taskId).Update("HelmRevision", helmRevision)
	if res.Error != nil || res.RowsAffected == 0 {
		return fmt.Errorf("UpdateEndpoint: nothing updated in AppServeAppTask with id %s", id)
	}

	return nil
}

func ConvertToPbAppServeApp(asa model.AppServeApp) *pb.AppServeApp {
	return &pb.AppServeApp{
		Id:              asa.ID.String(),
		Name:            asa.Name,
		ContractId:      asa.ContractId,
		TaskType:        asa.TaskType,
		EndpointUrl:     asa.EndpointUrl,
		TargetClusterId: asa.TargetClusterId,
		CreatedAt:       timestamppb.New(asa.CreatedAt),
		UpdatedAt:       timestamppb.New(asa.UpdatedAt),
	}
}

func ConvertToPbAppServeAppTask(task model.AppServeAppTask) *pb.AppServeAppTask {
	return &pb.AppServeAppTask{
		Id:           task.ID.String(),
		Version:      task.Version,
		Status:       task.Status,
		Output:       task.Output,
		ImageUrl:     task.ImageUrl,
		ArtifactUrl:  task.ArtifactUrl,
		Profile:      task.Profile,
		HelmRevision: task.HelmRevision,
		CreatedAt:    timestamppb.New(task.CreatedAt),
		UpdatedAt:    timestamppb.New(task.UpdatedAt),
	}
}