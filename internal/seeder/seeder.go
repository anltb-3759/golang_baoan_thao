package seeder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	fmt.Println("ðŸŒ± Starting database seeding...")

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

	fmt.Println("âœ… Database seeding completed successfully!")
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
			Name:         "Nguyá»…n VÄƒn An",
			Email:        "admin@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345678",
			Address:      "123 ÄÆ°á»ng LÃª Lá»£i, Quáº­n 1, TP. Há»“ ChÃ­ Minh",
			Role:         models.UserRoleSuperAdmin,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Tráº§n Thá»‹ BÃ¬nh",
			Email:        "manager@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345679",
			Address:      "456 ÄÆ°á»ng Nguyá»…n Huá»‡, Quáº­n 1, TP. Há»“ ChÃ­ Minh",
			Role:         models.UserRoleManager,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Pháº¡m Minh Tuáº¥n",
			Email:        "staff1@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345680",
			Address:      "789 ÄÆ°á»ng TÃ´n Äá»©c Tháº¯ng, Quáº­n 1, TP. Há»“ ChÃ­ Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "HoÃ ng Thá»‹ LiÃªn",
			Email:        "staff2@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345681",
			Address:      "321 ÄÆ°á»ng Äá»™c Láº­p, Quáº­n 1, TP. Há»“ ChÃ­ Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Le Quoc Viet",
			Email:        "staff3@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345684",
			Address:      "12 Tran Hung Dao, Quan 1, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Nguyen Gia Han",
			Email:        "staff4@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345685",
			Address:      "45 Vo Van Tan, Quan 3, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Tran Minh Khoa",
			Email:        "staff5@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345686",
			Address:      "88 Nguyen Dinh Chieu, Quan 1, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Do Thi Mai",
			Email:        "staff6@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345687",
			Address:      "19 Le Van Sy, Quan Phu Nhuan, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Pham Duc Long",
			Email:        "staff7@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345688",
			Address:      "201 Cach Mang Thang 8, Quan 10, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Bui Thanh Nhan",
			Email:        "staff8@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345689",
			Address:      "77 Cong Hoa, Quan Tan Binh, TP. Ho Chi Minh",
			Role:         models.UserRoleStaff,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "VÅ© ThÃ nh CÃ´ng",
			Email:        "citizen1@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345682",
			Address:      "555 ÄÆ°á»ng Láº¡c Long QuÃ¢n, Quáº­n 5, TP. Há»“ ChÃ­ Minh",
			Role:         models.UserRoleCitizen,
			Status:       models.UserStatusActive,
		},
		{
			Name:         "Äáº·ng Há»¯u Minh",
			Email:        "citizen2@example.com",
			PasswordHash: "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS",
			Phone:        "0912345683",
			Address:      "666 ÄÆ°á»ng Phan ÄÃ¬nh PhÃ¹ng, Quáº­n 3, TP. Há»“ ChÃ­ Minh",
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

	fmt.Println("âœ“ Users seeded")
	return nil
}

func seedDepartments(db *gorm.DB) error {
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}
	staff2, err := findUserByEmail(db, "staff2@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff2: %w", err)
	}

	departments := []models.Department{
		{
			Name:         "Phòng Cấp Giấy Tờ Tùy Thân",
			Code:         "PGTT",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff1.ID,
		},
		{
			Name:         "Phòng Đăng Ký Xe Cơ Giới",
			Code:         "PDXCG",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff2.ID,
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
		"name":            "Máº«u Ä‘Æ¡n cáº¥p CCCD",
		"fields":          []string{"full_name", "date_of_birth", "gender", "nationality"},
		"required_fields": []string{"full_name", "date_of_birth", "gender"},
	})
	formSchema2, _ := json.Marshal(map[string]interface{}{
		"name":            "Máº«u Ä‘Æ¡n cáº¥p Giáº¥y phÃ©p lÃ¡i xe",
		"fields":          []string{"license_class", "experience_years", "medical_exam_date"},
		"required_fields": []string{"license_class", "medical_exam_date"},
	})
	formSchema3, _ := json.Marshal(map[string]interface{}{
		"name":            "Máº«u Ä‘Æ¡n Ä‘Äƒng kÃ½ xe",
		"fields":          []string{"vehicle_type", "vehicle_brand", "chassis_no", "engine_no"},
		"required_fields": []string{"vehicle_type", "chassis_no", "engine_no"},
	})
	formSchema4, _ := json.Marshal(map[string]interface{}{
		"name":            "Máº«u Ä‘Æ¡n Ä‘Äƒng kÃ½ nháº­p há»c",
		"fields":          []string{"student_name", "date_of_birth", "school_name", "grade"},
		"required_fields": []string{"student_name", "date_of_birth", "school_name", "grade"},
	})
	formSchema5, _ := json.Marshal(map[string]interface{}{
		"name":            "Máº«u Ä‘Æ¡n Ä‘Äƒng kÃ½ khÃ¡m sá»©c khá»e",
		"fields":          []string{"full_name", "date_of_birth", "health_insurance_no", "preferred_date"},
		"required_fields": []string{"full_name", "date_of_birth"},
	})

	serviceTypes := []models.ServiceType{
		{
			Name:                    "Cáº¥p CCCD láº§n Ä‘áº§u",
			Code:                    "CCCD_NEW",
			Description:             "Cáº¥p CÄƒn cÆ°á»›c cÃ´ng dÃ¢n láº§n Ä‘áº§u cho cÃ´ng dÃ¢n Ä‘á»§ 14 tuá»•i",
			RequiredDocuments:       "Giáº¥y khai sinh, Chá»©ng minh thÆ° hoáº·c Há»™ chiáº¿u, áº¢nh mÃ u 3x4",
			FormSchema:              formSchema1,
			ProcessingTime:          intPtr(3),
			Fee:                     0.00,
			ResponsibleDepartmentID: &deptGTTT.ID,
			IsActive:                true,
		},
		{
			Name:                    "Cáº¥p Giáº¥y phÃ©p lÃ¡i xe háº¡ng A",
			Code:                    "LICENSE_CLASS_A",
			Description:             "Cáº¥p Giáº¥y phÃ©p lÃ¡i xe háº¡ng A (xe mÃ¡y)",
			RequiredDocuments:       "CCCD/Há»™ chiáº¿u, Giáº¥y chá»©ng nháº­n sá»©c khá»e, 4 áº£nh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(5),
			Fee:                     70000.00,
			ResponsibleDepartmentID: &deptXCG.ID,
			IsActive:                true,
		},
		{
			Name:                    "Cáº¥p Giáº¥y phÃ©p lÃ¡i xe háº¡ng C",
			Code:                    "LICENSE_CLASS_C",
			Description:             "Cáº¥p Giáº¥y phÃ©p lÃ¡i xe háº¡ng C (Ã´ tÃ´ nhá»)",
			RequiredDocuments:       "CCCD/Há»™ chiáº¿u, Giáº¥y chá»©ng nháº­n sá»©c khá»e, 4 áº£nh 3x4",
			FormSchema:              formSchema2,
			ProcessingTime:          intPtr(7),
			Fee:                     150000.00,
			ResponsibleDepartmentID: &deptXCG.ID,
			IsActive:                true,
		},
		{
			Name:              "ÄÄƒng kÃ½ xe mÃ¡y",
			Code:              "REGISTER_BIKE",
			Description:       "ÄÄƒng kÃ½ xe mÃ¡y táº¡i Cá»¥c ÄÄƒng kÃ½ LÃ¡i xe vÃ  Xe cÆ¡ giá»›i",
			RequiredDocuments: "HÃ³a Ä‘Æ¡n bÃ¡n hÃ ng, CCCD, Báº£ng kiá»ƒm tra ká»¹ thuáº­t",
			FormSchema:        formSchema3,
			ProcessingTime:    intPtr(1),
			Fee:               50000.00,
			IsActive:          true,
		},
		{
			Name:              "ÄÄƒng kÃ½ nháº­p há»c trÆ°á»ng cÃ´ng láº­p",
			Code:              "EDU_ENROLL",
			Description:       "ÄÄƒng kÃ½ nháº­p há»c cho há»c sinh vÃ o trÆ°á»ng tiá»ƒu há»c vÃ  trung há»c cÃ´ng láº­p",
			RequiredDocuments: "Giáº¥y khai sinh, Há»™ kháº©u hoáº·c Giáº¥y xÃ¡c nháº­n cÆ° trÃº, áº¢nh 3x4",
			FormSchema:        formSchema4,
			ProcessingTime:    intPtr(2),
			Fee:               0.00,
			IsActive:          true,
		},
		{
			Name:              "ÄÄƒng kÃ½ khÃ¡m sá»©c khá»e Ä‘á»‹nh ká»³",
			Code:              "HEALTH_CHECK",
			Description:       "ÄÄƒng kÃ½ dá»‹ch vá»¥ khÃ¡m sá»©c khá»e Ä‘á»‹nh ká»³ táº¡i cÆ¡ sá»Ÿ y táº¿ cÃ´ng láº­p",
			RequiredDocuments: "CCCD, Tháº» báº£o hiá»ƒm y táº¿",
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

	fmt.Println("âœ“ Service Types seeded")
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
			PermanentAddress:         "555 ÄÆ°á»ng Láº¡c Long QuÃ¢n, PhÆ°á»ng 9, Quáº­n 5, TP. Há»“ ChÃ­ Minh",
			EmailNotificationEnabled: true,
		},
		{
			UserID:                   citizen2.ID,
			CitizenIDNumber:          "987654321098",
			DateOfBirth:              &dob2,
			Gender:                   "Nam",
			PermanentAddress:         "666 ÄÆ°á»ng Phan ÄÃ¬nh PhÃ¹ng, PhÆ°á»ng 1, Quáº­n 3, TP. Há»“ ChÃ­ Minh",
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

	fmt.Println("âœ“ Citizen Profiles seeded")
	return nil
}

func seedStaffProfiles(db *gorm.DB) error {
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}
	staff2, err := findUserByEmail(db, "staff2@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff2: %w", err)
	}
	staff3, err := findUserByEmail(db, "staff3@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff3: %w", err)
	}
	staff4, err := findUserByEmail(db, "staff4@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff4: %w", err)
	}
	staff5, err := findUserByEmail(db, "staff5@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff5: %w", err)
	}
	staff6, err := findUserByEmail(db, "staff6@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff6: %w", err)
	}
	staff7, err := findUserByEmail(db, "staff7@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff7: %w", err)
	}
	staff8, err := findUserByEmail(db, "staff8@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff8: %w", err)
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
			UserID:       staff1.ID,
			DepartmentID: &deptGTTT.ID,
			Position:     "Trưởng phòng",
		},
		{
			UserID:       staff2.ID,
			DepartmentID: &deptXCG.ID,
			Position:     "Trưởng phòng",
		},
		{
			UserID:       staff3.ID,
			DepartmentID: &deptGTTT.ID,
			Position:     "Chuyên viên",
		},
		{
			UserID:       staff4.ID,
			DepartmentID: &deptGTTT.ID,
			Position:     "Chuyên viên",
		},
		{
			UserID:       staff5.ID,
			DepartmentID: &deptXCG.ID,
			Position:     "Chuyên viên",
		},
		{
			UserID:       staff6.ID,
			DepartmentID: &deptXCG.ID,
			Position:     "Chuyên viên",
		},
		{
			UserID:       staff7.ID,
			DepartmentID: nil,
			Position:     "Chuyên viên dự bị",
		},
		{
			UserID:       staff8.ID,
			DepartmentID: nil,
			Position:     "Chuyên viên dự bị",
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
	staff2, err := findUserByEmail(db, "staff2@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff2: %w", err)
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
			ApplicationCode:     "HCM-2026-001",
			CitizenUserID:       citizen1.ID,
			ServiceTypeID:       svcCCCD.ID,
			AssignedStaffUserID: &staff1.ID,
			Status:              models.ApplicationStatusProcessing,
			SubmittedData:       submittedData1,
			ResultNote:          "Đã tiếp nhận và đang xử lý hồ sơ.",
			SubmittedAt:         now.AddDate(0, 0, -7),
			ProcessingStartedAt: &yesterday,
		},
		{
			ApplicationCode: "HCM-2026-002",
			CitizenUserID:   citizen2.ID,
			ServiceTypeID:   svcLicenseA.ID,
			Status:          models.ApplicationStatusReceived,
			SubmittedData:   submittedData2,
			SubmittedAt:     now.AddDate(0, 0, -2),
		},
		{
			ApplicationCode:     "HCM-2026-003",
			CitizenUserID:       citizen1.ID,
			ServiceTypeID:       svcLicenseC.ID,
			AssignedStaffUserID: &staff1.ID,
			Status:              models.ApplicationStatusNeedMoreInfo,
			SubmittedData:       submittedData3,
			ResultNote:          "Vui lòng bổ sung giấy khám sức khỏe bản gốc.",
			SubmittedAt:         now.AddDate(0, 0, -4),
			ProcessingStartedAt: timePtr(now.AddDate(0, 0, -3)),
		},
		{
			ApplicationCode:     "HCM-2026-004",
			CitizenUserID:       citizen2.ID,
			ServiceTypeID:       svcLicenseC.ID,
			AssignedStaffUserID: &staff2.ID,
			Status:              models.ApplicationStatusApproved,
			SubmittedData:       submittedData2,
			ResultNote:          "Đã phê duyệt hồ sơ.",
			SubmittedAt:         now.AddDate(0, -1, 0),
			ProcessingStartedAt: timePtr(now.AddDate(0, -1, 5)),
			CompletedAt:         timePtr(now.AddDate(0, 0, -5)),
		},
		{
			ApplicationCode:     "HCM-2026-005",
			CitizenUserID:       citizen1.ID,
			ServiceTypeID:       svcCCCD.ID,
			AssignedStaffUserID: &staff2.ID,
			Status:              models.ApplicationStatusRejected,
			SubmittedData:       submittedData1,
			RejectedReason:      "Thông tin giấy tờ không khớp hồ sơ gốc.",
			SubmittedAt:         now.AddDate(0, 0, -10),
			ProcessingStartedAt: timePtr(now.AddDate(0, 0, -9)),
			CompletedAt:         timePtr(now.AddDate(0, 0, -8)),
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
			Title:         "Cáº­p nháº­t tráº¡ng thÃ¡i há»“ sÆ¡",
			Message:       "Há»“ sÆ¡ HCM-2024-001 cá»§a báº¡n Ä‘Ã£ chuyá»ƒn sang giai Ä‘oáº¡n xá»­ lÃ½. Vui lÃ²ng chá» káº¿t quáº£.",
			Type:          models.NotificationTypeReceived,
			IsRead:        false,
			CreatedAt:     time.Now(),
		},
		{
			UserID:    citizen2.ID,
			Title:     "ThÃ´ng bÃ¡o há»‡ thá»‘ng",
			Message:   "ChÃ o má»«ng báº¡n Ä‘áº¿n vá»›i Há»‡ thá»‘ng Quáº£n lÃ½ Dá»‹ch vá»¥ CÃ´ng. Báº¡n cÃ³ thá»ƒ Ä‘Äƒng kÃ½ cÃ¡c dá»‹ch vá»¥ cÃ´ng trá»±c tuyáº¿n.",
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

	fmt.Println("âœ“ Notifications seeded")
	return nil
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
