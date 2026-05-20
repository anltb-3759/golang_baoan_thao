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

func findUserByEmail(db *gorm.DB, email string) (models.User, error) {
	var u models.User
	err := db.Where("email = ?", email).First(&u).Error
	return u, err
}

func findDeptByCode(db *gorm.DB, code string) (models.Department, error) {
	var d models.Department
	err := db.Where("code = ?", code).First(&d).Error
	return d, err
}

func findServiceByCode(db *gorm.DB, code string) (models.ServiceType, error) {
	var s models.ServiceType
	err := db.Where("code = ?", code).First(&s).Error
	return s, err
}

func seedUsers(db *gorm.DB) error {
	users := []models.User{
		{
			Name:         "Nguyễn Văn An",
			Email:        "admin@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345678",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleSuperAdmin,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Trần Thị Bình",
			Email:        "manager@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345679",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleManager,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Phạm Minh Tuấn",
			Email:        "staff1@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345680",
			Address:      "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Hoàng Thị Liên",
			Email:        "staff2@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345681",
			Address:      "321 Đường Độc Lập, Quận 1, TP. Hồ Chí Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Vũ Thành Công",
			Email:        "citizen1@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345682",
			Address:      "555 Đường Lạc Long Quân, Quận 5, TP. Hồ Chí Minh",
			Role:         models.UserRoleCitizen,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Đặng Hữu Minh",
			Email:        "citizen2@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345683",
			Address:      "666 Đường Phan Đình Phùng, Quận 3, TP. Hồ Chí Minh",
			Role:         models.UserRoleCitizen,
			Status:       models.UserStatusActive,
		},
	}

	for i := range users {
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()
		if err := db.Where("email = ?", users[i].Email).FirstOrCreate(&users[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Users seeded")
	return nil
}

func seedDepartments(db *gorm.DB) error {
	manager, err := findUserByEmail(db, "manager@example.com")
	if err != nil {
		return fmt.Errorf("lookup manager: %w", err)
	}
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}

	departments := []models.Department{
		{
			Name:         "Phòng Cấp Giấy Tờ Tùy Thân",
			Code:         "PGTT",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &manager.ID,
		},
		{
			Name:         "Phòng Đăng Ký Xe Cơ Giới",
			Code:         "PDXCG",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff1.ID,
		},
		{
			Name:         "Phòng Cấp Giấy Phép Lái Xe",
			Code:         "PGPLX",
			Address:      "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: nil,
		},
	}

	for i := range departments {
		departments[i].CreatedAt = time.Now()
		departments[i].UpdatedAt = time.Now()
		if err := db.Where("code = ?", departments[i].Code).FirstOrCreate(&departments[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Departments seeded")
	return nil
}

func seedServiceTypes(db *gorm.DB) error {
	deptGTTT, err := findDeptByCode(db, "PGTT")
	if err != nil {
		return fmt.Errorf("lookup dept PGTT: %w", err)
	}
	deptXCG, err := findDeptByCode(db, "PDXCG")
	if err != nil {
		return fmt.Errorf("lookup dept PDXCG: %w", err)
	}

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
	formSchema4, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký nhập học",
		"fields":          []string{"student_name", "date_of_birth", "school_name", "grade"},
		"required_fields": []string{"student_name", "date_of_birth", "school_name", "grade"},
	})
	formSchema5, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký khám sức khỏe",
		"fields":          []string{"full_name", "date_of_birth", "health_insurance_no", "preferred_date"},
		"required_fields": []string{"full_name", "date_of_birth"},
	})

	serviceTypes := []models.ServiceType{
		{
			Name:                    "Cấp CCCD lần đầu",
			Code:                    "CCCD_NEW",
			Category:                models.ServiceCategoryAdministrative,
			Description:             "Cấp Căn cước công dân lần đầu cho công dân đủ 14 tuổi",
			RequiredDocuments:       "Giấy khai sinh, Chứng minh thư hoặc Hộ chiếu, Ảnh màu 3x4",
			FormSchema:              formSchema1,
			ProcessingTime:          intPtr(3),
			Fee:                     0.00,
			ResponsibleDepartmentID: &deptGTTT.ID,
			IsActive:                true,
		},
		{
			Name:                    "Cấp Giấy phép lái xe hạng A",
			Code:                    "LICENSE_CLASS_A",
			Category:                models.ServiceCategoryAdministrative,
			Description:             "Cấp Giấy phép lái xe hạng A (xe máy)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(5),
			Fee:                     70000.00,
			ResponsibleDepartmentID: &deptXCG.ID,
			IsActive:                true,
		},
		{
			Name:                    "Cấp Giấy phép lái xe hạng C",
			Code:                    "LICENSE_CLASS_C",
			Category:                models.ServiceCategoryAdministrative,
			Description:             "Cấp Giấy phép lái xe hạng C (ô tô nhỏ)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(7),
			Fee:                     150000.00,
			ResponsibleDepartmentID: &deptXCG.ID,
			IsActive:                true,
		},
		{
			Name:              "Đăng ký xe máy",
			Code:              "REGISTER_BIKE",
			Category:          models.ServiceCategoryAdministrative,
			Description:       "Đăng ký xe máy tại Cục Đăng ký Lái xe và Xe cơ giới",
			RequiredDocuments: "Hóa đơn bán hàng, CCCD, Bảng kiểm tra kỹ thuật",
			FormSchema:        formSchema3,
			ProcessingTime:    intPtr(1),
			Fee:               50000.00,
			IsActive:          true,
		},
		{
			Name:              "Đăng ký nhập học trường công lập",
			Code:              "EDU_ENROLL",
			Category:          models.ServiceCategoryEducation,
			Description:       "Đăng ký nhập học cho học sinh vào trường tiểu học và trung học công lập",
			RequiredDocuments: "Giấy khai sinh, Hộ khẩu hoặc Giấy xác nhận cư trú, Ảnh 3x4",
			FormSchema:        formSchema4,
			ProcessingTime:    intPtr(2),
			Fee:               0.00,
			IsActive:          true,
		},
		{
			Name:              "Đăng ký khám sức khỏe định kỳ",
			Code:              "HEALTH_CHECK",
			Category:          models.ServiceCategoryHealth,
			Description:       "Đăng ký dịch vụ khám sức khỏe định kỳ tại cơ sở y tế công lập",
			RequiredDocuments: "CCCD, Thẻ bảo hiểm y tế",
			FormSchema:        formSchema5,
			ProcessingTime:    intPtr(1),
			Fee:               30000.00,
			IsActive:          true,
		},
	}

	for i := range serviceTypes {
		serviceTypes[i].CreatedAt = time.Now()
		serviceTypes[i].UpdatedAt = time.Now()
		if err := db.Where("code = ?", serviceTypes[i].Code).FirstOrCreate(&serviceTypes[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Service Types seeded")
	return nil
}

func seedCitizenProfiles(db *gorm.DB) error {
	citizen1, err := findUserByEmail(db, "citizen1@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen1: %w", err)
	}
	citizen2, err := findUserByEmail(db, "citizen2@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen2: %w", err)
	}

	dob1 := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	dob2 := time.Date(1985, 10, 20, 0, 0, 0, 0, time.UTC)

	profiles := []models.CitizenProfile{
		{
			UserID:                   citizen1.ID,
			CitizenIDNumber:          "123456789012",
			DateOfBirth:              &dob1,
			Gender:                   "Nam",
			PermanentAddress:         "555 Đường Lạc Long Quân, Phường 9, Quận 5, TP. Hồ Chí Minh",
			EmailNotificationEnabled: true,
		},
		{
			UserID:                   citizen2.ID,
			CitizenIDNumber:          "987654321098",
			DateOfBirth:              &dob2,
			Gender:                   "Nam",
			PermanentAddress:         "666 Đường Phan Đình Phùng, Phường 1, Quận 3, TP. Hồ Chí Minh",
			EmailNotificationEnabled: true,
		},
	}

	for i := range profiles {
		profiles[i].CreatedAt = time.Now()
		profiles[i].UpdatedAt = time.Now()
		if err := db.Where("user_id = ?", profiles[i].UserID).FirstOrCreate(&profiles[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Citizen Profiles seeded")
	return nil
}

func seedStaffProfiles(db *gorm.DB) error {
	manager, err := findUserByEmail(db, "manager@example.com")
	if err != nil {
		return fmt.Errorf("lookup manager: %w", err)
	}
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}
	staff2, err := findUserByEmail(db, "staff2@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff2: %w", err)
	}
	deptGTTT, err := findDeptByCode(db, "PGTT")
	if err != nil {
		return fmt.Errorf("lookup dept PGTT: %w", err)
	}
	deptXCG, err := findDeptByCode(db, "PDXCG")
	if err != nil {
		return fmt.Errorf("lookup dept PDXCG: %w", err)
	}

	profiles := []models.StaffProfile{
		{
			UserID:       manager.ID,
			DepartmentID: &deptGTTT.ID,
			Position:     "Trưởng phòng",
		},
		{
			UserID:       staff1.ID,
			DepartmentID: &deptGTTT.ID,
			Position:     "Chuyên viên cao cấp",
		},
		{
			UserID:       staff2.ID,
			DepartmentID: &deptXCG.ID,
			Position:     "Chuyên viên",
		},
	}

	for i := range profiles {
		profiles[i].CreatedAt = time.Now()
		profiles[i].UpdatedAt = time.Now()
		if err := db.Where("user_id = ?", profiles[i].UserID).FirstOrCreate(&profiles[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Staff Profiles seeded")
	return nil
}

func seedApplications(db *gorm.DB) error {
	citizen1, err := findUserByEmail(db, "citizen1@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen1: %w", err)
	}
	citizen2, err := findUserByEmail(db, "citizen2@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen2: %w", err)
	}
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}
	svcCCCD, err := findServiceByCode(db, "CCCD_NEW")
	if err != nil {
		return fmt.Errorf("lookup service CCCD_NEW: %w", err)
	}
	svcLicenseA, err := findServiceByCode(db, "LICENSE_CLASS_A")
	if err != nil {
		return fmt.Errorf("lookup service LICENSE_CLASS_A: %w", err)
	}
	svcLicenseC, err := findServiceByCode(db, "LICENSE_CLASS_C")
	if err != nil {
		return fmt.Errorf("lookup service LICENSE_CLASS_C: %w", err)
	}

	submittedData1, _ := json.Marshal(map[string]interface{}{
		"full_name":      "Vũ Thành Công",
		"date_of_birth":  "1990-05-15",
		"gender":         "Nam",
		"nationality":    "Việt Nam",
		"home_address":   "555 Đường Lạc Long Quân, Quận 5, TP. HCM",
		"permanent_addr": "555 Đường Lạc Long Quân, Quận 5, TP. HCM",
	})
	submittedData2, _ := json.Marshal(map[string]interface{}{
		"full_name":          "Đặng Hữu Minh",
		"license_class":      "A1",
		"experience_years":   3,
		"medical_exam_date":  "2024-05-10",
		"health_condition":   "Good",
		"driving_experience": "Bằng lái xe máy 3 năm",
	})
	submittedData3, _ := json.Marshal(map[string]interface{}{
		"full_name":          "Vũ Thành Công",
		"license_class":      "C",
		"experience_years":   5,
		"medical_exam_date":  "2024-05-15",
		"health_condition":   "Good",
		"driving_experience": "Bằng lái xe máy 5 năm, muốn nâng cấp bằng C",
	})

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	applications := []models.Application{
		{
			ApplicationCode:     "HCM-2024-001",
			CitizenUserID:       citizen1.ID,
			ServiceTypeID:       svcCCCD.ID,
			AssignedStaffUserID: &staff1.ID,
			Status:              models.ApplicationStatusProcessing,
			SubmittedData:       submittedData1,
			ResultNote:          "Đã kiểm tra giấy tờ, đang xử lý cấp CCCD lần đầu",
			SubmittedAt:         now.AddDate(0, 0, -7),
			ProcessingStartedAt: &yesterday,
		},
		{
			ApplicationCode: "HCM-2024-002",
			CitizenUserID:   citizen2.ID,
			ServiceTypeID:   svcLicenseA.ID,
			Status:          models.ApplicationStatusReceived,
			SubmittedData:   submittedData2,
			SubmittedAt:     now.AddDate(0, 0, -2),
		},
		{
			ApplicationCode:     "HCM-2024-003",
			CitizenUserID:       citizen1.ID,
			ServiceTypeID:       svcLicenseC.ID,
			AssignedStaffUserID: &staff1.ID,
			Status:              models.ApplicationStatusApproved,
			SubmittedData:       submittedData3,
			ResultNote:          "Đã phê duyệt. Giấy phép lái xe hạng C sẵn sàng lấy tại cơ quan.",
			SubmittedAt:         now.AddDate(0, -1, 0),
			ProcessingStartedAt: timePtr(now.AddDate(0, -1, 5)),
			CompletedAt:         timePtr(now.AddDate(0, 0, -5)),
		},
	}

	for i := range applications {
		applications[i].CreatedAt = time.Now()
		applications[i].UpdatedAt = time.Now()
		if err := db.Where("application_code = ?", applications[i].ApplicationCode).FirstOrCreate(&applications[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Applications seeded")
	return nil
}

func seedNotifications(db *gorm.DB) error {
	citizen1, err := findUserByEmail(db, "citizen1@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen1: %w", err)
	}
	citizen2, err := findUserByEmail(db, "citizen2@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen2: %w", err)
	}

	var app models.Application
	db.Where("application_code = ?", "HCM-2024-001").First(&app)

	notifications := []models.Notification{
		{
			UserID:        citizen1.ID,
			ApplicationID: &app.ID,
			Title:         "Cập nhật trạng thái hồ sơ",
			Message:       "Hồ sơ HCM-2024-001 của bạn đã chuyển sang giai đoạn xử lý. Vui lòng chờ kết quả.",
			Type:          models.NotificationTypeReceived,
			IsRead:        false,
			CreatedAt:     time.Now(),
		},
		{
			UserID:    citizen2.ID,
			Title:     "Thông báo hệ thống",
			Message:   "Chào mừng bạn đến với Hệ thống Quản lý Dịch vụ Công. Bạn có thể đăng ký các dịch vụ công trực tuyến.",
			Type:      models.NotificationTypeSystem,
			IsRead:    true,
			ReadAt:    timePtr(time.Now().Add(-2 * time.Hour)),
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}

	for i := range notifications {
		if err := db.Where("user_id = ? AND title = ?", notifications[i].UserID, notifications[i].Title).
			FirstOrCreate(&notifications[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Notifications seeded")
	return nil
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
