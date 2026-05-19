package seeder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	fmt.Println("🌱 Starting database seeding...")

	if err := seedUsers(db); err != nil {
		return fmt.Errorf("error seeding users: %w", err)
	}

	if err := seedDepartments(db); err != nil {
		return fmt.Errorf("error seeding departments: %w", err)
	}

	if err := seedServiceTypes(db); err != nil {
		return fmt.Errorf("error seeding service types: %w", err)
	}

	if err := seedCitizenProfiles(db); err != nil {
		return fmt.Errorf("error seeding citizen profiles: %w", err)
	}

	if err := seedStaffProfiles(db); err != nil {
		return fmt.Errorf("error seeding staff profiles: %w", err)
	}

	if err := seedApplications(db); err != nil {
		return fmt.Errorf("error seeding applications: %w", err)
	}

	if err := seedNotifications(db); err != nil {
		return fmt.Errorf("error seeding notifications: %w", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
	return nil
}

func seedUsers(db *gorm.DB) error {
	users := []models.User{
		{
			ID:           "00000000-0000-0000-0000-000000000001",
			Name:         "Nguyễn Văn An",
			Email:        "admin@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2", // hash of "admin123"
			Phone:        "0912345678",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleSuperAdmin,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "00000000-0000-0000-0000-000000000002",
			Name:         "Trần Thị Bình",
			Email:        "manager@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2",
			Phone:        "0912345679",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleManager,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "00000000-0000-0000-0000-000000000003",
			Name:         "Phạm Minh Tuấn",
			Email:        "staff1@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2",
			Phone:        "0912345680",
			Address:      "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "00000000-0000-0000-0000-000000000004",
			Name:         "Hoàng Thị Liên",
			Email:        "staff2@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2",
			Phone:        "0912345681",
			Address:      "321 Đường Độc Lập, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "00000000-0000-0000-0000-000000000005",
			Name:         "Vũ Thành Công",
			Email:        "citizen1@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2",
			Phone:        "0912345682",
			Address:      "555 Đường Lạc Long Quân, Quận 5, TP. Hồ Chí Minh",
			Role:         models.UserRoleCitizen,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "00000000-0000-0000-0000-000000000006",
			Name:         "Đặng Hữu Minh",
			Email:        "citizen2@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcg7b3XeKeUxWdeS86E36P4/TVm2",
			Phone:        "0912345683",
			Address:      "666 Đường Phan Đình Phùng, Quận 3, TP. Hồ Chí Minh",
			Role:         models.UserRoleCitizen,
			Status:       models.UserStatusActive,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, user := range users {
		if err := db.Where("email = ?", user.Email).FirstOrCreate(&user).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Users seeded")
	return nil
}

func seedDepartments(db *gorm.DB) error {
	leaderID2 := "00000000-0000-0000-0000-000000000002"
	leaderID3 := "00000000-0000-0000-0000-000000000003"

	departments := []models.Department{
		{
			ID:           "10000000-0000-0000-0000-000000000001",
			Name:         "Phòng Cấp Giấy Tờ Tùy Thân",
			Code:         "PGTT",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &leaderID2,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "10000000-0000-0000-0000-000000000002",
			Name:         "Phòng Đăng Ký Xe Cơ Giới",
			Code:         "PDXCG",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &leaderID3,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "10000000-0000-0000-0000-000000000003",
			Name:         "Phòng Cấp Giấy Phép Lái Xe",
			Code:         "PGPLX",
			Address:      "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: nil,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, dept := range departments {
		if err := db.Where("code = ?", dept.Code).FirstOrCreate(&dept).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Departments seeded")
	return nil
}

func seedServiceTypes(db *gorm.DB) error {
	deptID1 := "10000000-0000-0000-0000-000000000001"
	deptID2 := "10000000-0000-0000-0000-000000000002"

	formSchema1, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn cấp CCCD",
		"fields":          []string{"full_name", "date_of_birth", "gender", "nationality"},
		"required_fields": []string{"full_name", "date_of_birth", "gender"},
	})

	formSchema2, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn cấp Giấy phép lái xe",
		"fields":          []string{"license_class", "experience_years", "medical_exam_date"},
		"required_fields": []string{"license_class", "medical_exam_date"},
	})

	formSchema3, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký xe",
		"fields":          []string{"vehicle_type", "vehicle_brand", "chassis_no", "engine_no"},
		"required_fields": []string{"vehicle_type", "chassis_no", "engine_no"},
	})

	serviceTypes := []models.ServiceType{
		{
			ID:                      "20000000-0000-0000-0000-000000000001",
			Name:                    "Cấp CCCD lần đầu",
			Code:                    "CCCD_NEW",
			Description:             "Cấp Căn cước công dân lần đầu cho công dân đủ 14 tuổi",
			RequiredDocuments:       "Giấy khai sinh, Chứng minh thư hoặc Hộ chiếu, Ảnh màu 3x4",
			FormSchema:              formSchema1,
			ProcessingTime:          intPtr(3),
			Fee:                     0.00,
			ResponsibleDepartmentID: &deptID1,
			IsActive:                true,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		},
		{
			ID:                      "20000000-0000-0000-0000-000000000002",
			Name:                    "Cấp Giấy phép lái xe hạng A",
			Code:                    "LICENSE_CLASS_A",
			Description:             "Cấp Giấy phép lái xe hạng A (xe máy)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(5),
			Fee:                     70000.00,
			ResponsibleDepartmentID: &deptID2,
			IsActive:                true,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		},
		{
			ID:                      "20000000-0000-0000-0000-000000000003",
			Name:                    "Cấp Giấy phép lái xe hạng C",
			Code:                    "LICENSE_CLASS_C",
			Description:             "Cấp Giấy phép lái xe hạng C (ô tô nhỏ)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(7),
			Fee:                     150000.00,
			ResponsibleDepartmentID: &deptID2,
			IsActive:                true,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		},
		{
			ID:                      "20000000-0000-0000-0000-000000000004",
			Name:                    "Đăng ký xe máy",
			Code:                    "REGISTER_BIKE",
			Description:             "Đăng ký xe máy tại Cục Đăng ký Lái xe và Xe cơ giới",
			RequiredDocuments:       "Hóa đơn bán hàng, CCCD, Bảng kiểm tra kỹ thuật",
			FormSchema:              formSchema3,
			ProcessingTime:          intPtr(1),
			Fee:                     50000.00,
			ResponsibleDepartmentID: nil,
			IsActive:                true,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		},
	}

	for _, st := range serviceTypes {
		if err := db.Where("code = ?", st.Code).FirstOrCreate(&st).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Service Types seeded")
	return nil
}

func seedCitizenProfiles(db *gorm.DB) error {
	dob1 := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dob2 := time.Date(1985, 10, 20, 0, 0, 0, 0, time.UTC)

	profiles := []models.CitizenProfile{
		{
			ID:                       "30000000-0000-0000-0000-000000000001",
			UserID:                   "00000000-0000-0000-0000-000000000005",
			CitizenIDNumber:          "123456789012",
			DateOfBirth:              &dob1,
			Gender:                   "Nam",
			PermanentAddress:         "555 Đường Lạc Long Quân, Phường 9, Quận 5, TP. Hồ Chí Minh",
			EmailNotificationEnabled: true,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		},
		{
			ID:                       "30000000-0000-0000-0000-000000000002",
			UserID:                   "00000000-0000-0000-0000-000000000006",
			CitizenIDNumber:          "987654321098",
			DateOfBirth:              &dob2,
			Gender:                   "Nam",
			PermanentAddress:         "666 Đường Phan Đình Phùng, Phường 1, Quận 3, TP. Hồ Chí Minh",
			EmailNotificationEnabled: true,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		},
	}

	for _, profile := range profiles {
		if err := db.Where("user_id = ?", profile.UserID).FirstOrCreate(&profile).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Citizen Profiles seeded")
	return nil
}

func seedStaffProfiles(db *gorm.DB) error {
	deptID1 := "10000000-0000-0000-0000-000000000001"
	deptID2 := "10000000-0000-0000-0000-000000000002"

	profiles := []models.StaffProfile{
		{
			ID:           "40000000-0000-0000-0000-000000000001",
			UserID:       "00000000-0000-0000-0000-000000000002",
			DepartmentID: &deptID1,
			Position:     "Department Manager",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "40000000-0000-0000-0000-000000000002",
			UserID:       "00000000-0000-0000-0000-000000000003",
			DepartmentID: &deptID1,
			Position:     "Senior Staff",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "40000000-0000-0000-0000-000000000003",
			UserID:       "00000000-0000-0000-0000-000000000004",
			DepartmentID: &deptID2,
			Position:     "Junior Staff",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	for _, profile := range profiles {
		if err := db.Where("user_id = ?", profile.UserID).FirstOrCreate(&profile).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Staff Profiles seeded")
	return nil
}

func seedApplications(db *gorm.DB) error {
	staffID := "00000000-0000-0000-0000-000000000003"

	submittedData1, _ := json.Marshal(map[string]interface{}{
		"full_name":      "Vũ Thành Công",
		"date_of_birth":  "1990-05-15",
		"gender":         "Nam",
		"nationality":    "Việt Nam",
		"home_address":   "555 Đường Lạc Long Quân, Quận 5, TP. HCM",
		"permanent_addr": "555 Đường Lạc Long Quân, Quận 5, TP. HCM",
	})

	submittedData2, _ := json.Marshal(map[string]interface{}{
		"full_name":           "Đặng Hữu Minh",
		"license_class":       "A1",
		"experience_years":    3,
		"medical_exam_date":   "2024-05-10",
		"health_condition":    "Good",
		"driving_experience": "Bằng lái xe máy 3 năm",
	})

	submittedData3, _ := json.Marshal(map[string]interface{}{
		"full_name":           "Vũ Thành Công",
		"license_class":       "C",
		"experience_years":    5,
		"medical_exam_date":   "2024-05-15",
		"health_condition":    "Good",
		"driving_experience": "Bằng lái xe máy 5 năm, muốn nâng cấp bằng C",
	})

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	weekAgo := now.AddDate(0, 0, -7)

	applications := []models.Application{
		{
			ID:                  "50000000-0000-0000-0000-000000000001",
			ApplicationCode:     "HCM-2024-001",
			CitizenUserID:       "00000000-0000-0000-0000-000000000005",
			ServiceTypeID:       "20000000-0000-0000-0000-000000000001",
			AssignedStaffUserID: &staffID,
			Status:              models.ApplicationStatusProcessing,
			SubmittedData:       submittedData1,
			ResultNote:          "Đã kiểm tra giấy tờ, đang xử lý cấp CCCD lần đầu",
			SubmittedAt:         weekAgo,
			ProcessingStartedAt: &yesterday,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
		{
			ID:                  "50000000-0000-0000-0000-000000000002",
			ApplicationCode:     "HCM-2024-002",
			CitizenUserID:       "00000000-0000-0000-0000-000000000006",
			ServiceTypeID:       "20000000-0000-0000-0000-000000000002",
			AssignedStaffUserID: nil,
			Status:              models.ApplicationStatusReceived,
			SubmittedData:       submittedData2,
			ResultNote:          "",
			SubmittedAt:         now.AddDate(0, 0, -2),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
		{
			ID:                  "50000000-0000-0000-0000-000000000003",
			ApplicationCode:     "HCM-2024-003",
			CitizenUserID:       "00000000-0000-0000-0000-000000000005",
			ServiceTypeID:       "20000000-0000-0000-0000-000000000003",
			AssignedStaffUserID: &staffID,
			Status:              models.ApplicationStatusApproved,
			SubmittedData:       submittedData3,
			ResultNote:          "Đã phê duyệt. Giấy phép lái xe hạng C sẵn sàng lấy tại cơ quan.",
			SubmittedAt:         time.Now().AddDate(0, -1, 0),
			ProcessingStartedAt: timePtr(now.AddDate(0, -1, 5)),
			CompletedAt:         timePtr(now.AddDate(0, 0, -5)),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
	}

	for _, app := range applications {
		if err := db.Where("application_code = ?", app.ApplicationCode).FirstOrCreate(&app).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Applications seeded")
	return nil
}

func seedNotifications(db *gorm.DB) error {
	appID := "50000000-0000-0000-0000-000000000001"

	notifications := []models.Notification{
		{
			ID:            "60000000-0000-0000-0000-000000000001",
			UserID:        "00000000-0000-0000-0000-000000000005",
			ApplicationID: &appID,
			Title:         "Cập nhật trạng thái hồ sơ",
			Message:       "Hồ sơ HCM-2024-001 của bạn đã chuyển sang giai đoạn xử lý. Vui lòng chờ kết quả.",
			Type:          models.NotificationTypeReceived,
			IsRead:        false,
			CreatedAt:     time.Now(),
		},
		{
			ID:            "60000000-0000-0000-0000-000000000002",
			UserID:        "00000000-0000-0000-0000-000000000006",
			ApplicationID: nil,
			Title:         "Thông báo hệ thống",
			Message:       "Chào mừng bạn đến với Hệ thống Quản lý Dịch vụ Công. Bạn có thể đăng ký các dịch vụ công trực tuyến.",
			Type:          models.NotificationTypeSystem,
			IsRead:        true,
			ReadAt:        timePtr(time.Now().Add(-2 * time.Hour)),
			CreatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	for _, notif := range notifications {
		if err := db.Where("id = ?", notif.ID).FirstOrCreate(&notif).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Notifications seeded")
	return nil
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
