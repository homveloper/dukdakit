package diffit

import (
	"testing"
	"time"
)

// Benchmark structures from simple to complex

// BenchUser represents a basic user structure for benchmarking
type BenchUser struct {
	ID       int    `bson:"id"`
	Name     string `bson:"name"`
	Email    string `bson:"email"`
	Age      int    `bson:"age"`
	IsActive bool   `bson:"is_active"`
}

// MediumUser represents a user with more fields and nested data
type MediumUser struct {
	ID        int       `bson:"id"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Age       int       `bson:"age"`
	IsActive  bool      `bson:"is_active"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	
	// Nested profile
	Profile UserProfile `bson:"profile"`
	
	// Arrays
	Tags       []string `bson:"tags"`
	Scores     []int    `bson:"scores"`
	
	// Map
	Metadata map[string]interface{} `bson:"metadata"`
}

type UserProfile struct {
	FirstName   string  `bson:"first_name"`
	LastName    string  `bson:"last_name"`
	Bio         string  `bson:"bio"`
	Avatar      string  `bson:"avatar"`
	Phone       *string `bson:"phone,omitempty"`
	Country     string  `bson:"country"`
	Timezone    string  `bson:"timezone"`
}

// ComplexUser represents a highly complex user structure
type ComplexUser struct {
	ID        int       `bson:"id"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Age       int       `bson:"age"`
	IsActive  bool      `bson:"is_active"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	
	// Complex nested structures
	Profile        UserProfile        `bson:"profile"`
	Settings       UserSettings       `bson:"settings"`
	Preferences    UserPreferences    `bson:"preferences"`
	Subscriptions  []Subscription     `bson:"subscriptions"`
	Permissions    []Permission       `bson:"permissions"`
	SocialAccounts []SocialAccount    `bson:"social_accounts"`
	
	// Complex arrays
	Tags           []string                   `bson:"tags"`
	Scores         []int                      `bson:"scores"`
	ActivityLogs   []ActivityLog              `bson:"activity_logs"`
	CustomFields   []CustomField              `bson:"custom_fields"`
	Addresses      []Address                  `bson:"addresses"`
	
	// Complex maps
	Metadata       map[string]interface{}     `bson:"metadata"`
	Features       map[string]bool            `bson:"features"`
	Configurations map[string]Configuration   `bson:"configurations"`
	
	// Pointers to complex structures
	Company        *Company                   `bson:"company,omitempty"`
	Manager        *ComplexUser               `bson:"manager,omitempty"`
	
	// Time-related fields
	LastLoginAt    *time.Time                 `bson:"last_login_at,omitempty"`
	PasswordSetAt  time.Time                  `bson:"password_set_at"`
	SessionTimeout time.Duration              `bson:"session_timeout"`
}

type UserSettings struct {
	Theme           string            `bson:"theme"`
	Language        string            `bson:"language"`
	NotificationsOn bool              `bson:"notifications_on"`
	EmailFrequency  string            `bson:"email_frequency"`
	Privacy         PrivacySettings   `bson:"privacy"`
	Security        SecuritySettings  `bson:"security"`
}

type PrivacySettings struct {
	ProfileVisible   bool `bson:"profile_visible"`
	EmailVisible     bool `bson:"email_visible"`
	LastSeenVisible  bool `bson:"last_seen_visible"`
	OnlineStatus     bool `bson:"online_status"`
	SearchIndexable  bool `bson:"search_indexable"`
}

type SecuritySettings struct {
	TwoFactorEnabled    bool      `bson:"two_factor_enabled"`
	PasswordChangeAt    time.Time `bson:"password_change_at"`
	LoginAttempts       int       `bson:"login_attempts"`
	LastFailedLoginAt   *time.Time `bson:"last_failed_login_at,omitempty"`
	AllowedIPs          []string  `bson:"allowed_ips"`
	BlockedIPs          []string  `bson:"blocked_ips"`
}

type UserPreferences struct {
	Dashboard        DashboardPrefs `bson:"dashboard"`
	Notifications    NotificationPrefs `bson:"notifications"`
	Appearance       AppearancePrefs `bson:"appearance"`
	Integrations     []string `bson:"integrations"`
}

type DashboardPrefs struct {
	Layout          string   `bson:"layout"`
	Widgets         []string `bson:"widgets"`
	DefaultView     string   `bson:"default_view"`
	RefreshInterval int      `bson:"refresh_interval"`
}

type NotificationPrefs struct {
	Email    bool `bson:"email"`
	SMS      bool `bson:"sms"`
	Push     bool `bson:"push"`
	InApp    bool `bson:"in_app"`
	Digest   bool `bson:"digest"`
}

type AppearancePrefs struct {
	ColorScheme    string `bson:"color_scheme"`
	FontSize       string `bson:"font_size"`
	CompactMode    bool   `bson:"compact_mode"`
	AnimationsOn   bool   `bson:"animations_on"`
}

type Subscription struct {
	ID          string            `bson:"id"`
	PlanName    string            `bson:"plan_name"`
	Status      string            `bson:"status"`
	StartDate   time.Time         `bson:"start_date"`
	EndDate     *time.Time        `bson:"end_date,omitempty"`
	Amount      float64           `bson:"amount"`
	Currency    string            `bson:"currency"`
	Features    []string          `bson:"features"`
	Metadata    map[string]string `bson:"metadata"`
}

type Permission struct {
	Resource string   `bson:"resource"`
	Actions  []string `bson:"actions"`
	Scope    string   `bson:"scope"`
	Granted  bool     `bson:"granted"`
	GrantedAt time.Time `bson:"granted_at"`
	GrantedBy string   `bson:"granted_by"`
}

type SocialAccount struct {
	Provider    string            `bson:"provider"`
	ExternalID  string            `bson:"external_id"`
	Username    string            `bson:"username"`
	DisplayName string            `bson:"display_name"`
	Email       *string           `bson:"email,omitempty"`
	Avatar      *string           `bson:"avatar,omitempty"`
	IsVerified  bool              `bson:"is_verified"`
	ConnectedAt time.Time         `bson:"connected_at"`
	LastSyncAt  *time.Time        `bson:"last_sync_at,omitempty"`
	Metadata    map[string]string `bson:"metadata"`
}

type ActivityLog struct {
	ID          string                 `bson:"id"`
	Action      string                 `bson:"action"`
	Resource    string                 `bson:"resource"`
	ResourceID  string                 `bson:"resource_id"`
	Timestamp   time.Time              `bson:"timestamp"`
	IPAddress   string                 `bson:"ip_address"`
	UserAgent   string                 `bson:"user_agent"`
	Result      string                 `bson:"result"`
	Duration    time.Duration          `bson:"duration"`
	Details     map[string]interface{} `bson:"details"`
}

type CustomField struct {
	Name        string      `bson:"name"`
	Type        string      `bson:"type"`
	Value       interface{} `bson:"value"`
	IsRequired  bool        `bson:"is_required"`
	IsSearchable bool       `bson:"is_searchable"`
	Category    string      `bson:"category"`
	CreatedAt   time.Time   `bson:"created_at"`
	UpdatedAt   time.Time   `bson:"updated_at"`
}

type Address struct {
	Type         string  `bson:"type"`
	Street       string  `bson:"street"`
	City         string  `bson:"city"`
	State        string  `bson:"state"`
	Country      string  `bson:"country"`
	PostalCode   string  `bson:"postal_code"`
	Latitude     *float64 `bson:"latitude,omitempty"`
	Longitude    *float64 `bson:"longitude,omitempty"`
	IsPrimary    bool    `bson:"is_primary"`
	IsVerified   bool    `bson:"is_verified"`
}

type Configuration struct {
	Enabled   bool                   `bson:"enabled"`
	Settings  map[string]interface{} `bson:"settings"`
	UpdatedAt time.Time              `bson:"updated_at"`
	UpdatedBy string                 `bson:"updated_by"`
}

type Company struct {
	ID          string            `bson:"id"`
	Name        string            `bson:"name"`
	Industry    string            `bson:"industry"`
	Size        string            `bson:"size"`
	Website     string            `bson:"website"`
	Founded     *time.Time        `bson:"founded,omitempty"`
	Address     Address           `bson:"address"`
	Metadata    map[string]string `bson:"metadata"`
}

// Benchmark helper functions

func createSimpleUser() BenchUser {
	return BenchUser{
		ID:       1,
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		IsActive: true,
	}
}

func createMediumUser() MediumUser {
	now := time.Now()
	phone := "+1-555-1234"
	return MediumUser{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Profile: UserProfile{
			FirstName: "John",
			LastName:  "Doe",
			Bio:       "Software Engineer",
			Avatar:    "https://example.com/avatar.jpg",
			Phone:     &phone,
			Country:   "US",
			Timezone:  "America/New_York",
		},
		Tags:   []string{"developer", "golang", "mongodb"},
		Scores: []int{95, 87, 92, 88},
		Metadata: map[string]interface{}{
			"source":     "api",
			"version":    "1.0",
			"experiment": true,
			"score":      95.5,
		},
	}
}

func createComplexUser() ComplexUser {
	now := time.Now()
	phone := "+1-555-1234"
	lastLogin := now.Add(-2 * time.Hour)
	lat := 40.7128
	lng := -74.0060
	founded := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	
	return ComplexUser{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		
		Profile: UserProfile{
			FirstName: "John",
			LastName:  "Doe",
			Bio:       "Senior Software Engineer specializing in distributed systems",
			Avatar:    "https://example.com/avatar.jpg",
			Phone:     &phone,
			Country:   "US",
			Timezone:  "America/New_York",
		},
		
		Settings: UserSettings{
			Theme:           "dark",
			Language:        "en",
			NotificationsOn: true,
			EmailFrequency:  "daily",
			Privacy: PrivacySettings{
				ProfileVisible:  true,
				EmailVisible:    false,
				LastSeenVisible: true,
				OnlineStatus:    true,
				SearchIndexable: true,
			},
			Security: SecuritySettings{
				TwoFactorEnabled:  true,
				PasswordChangeAt:  now.Add(-30 * 24 * time.Hour),
				LoginAttempts:     0,
				LastFailedLoginAt: nil,
				AllowedIPs:        []string{"192.168.1.0/24", "10.0.0.0/8"},
				BlockedIPs:        []string{"192.168.100.1"},
			},
		},
		
		Preferences: UserPreferences{
			Dashboard: DashboardPrefs{
				Layout:          "grid",
				Widgets:         []string{"metrics", "charts", "logs", "alerts"},
				DefaultView:     "overview",
				RefreshInterval: 30,
			},
			Notifications: NotificationPrefs{
				Email: true,
				SMS:   false,
				Push:  true,
				InApp: true,
				Digest: true,
			},
			Appearance: AppearancePrefs{
				ColorScheme:  "auto",
				FontSize:     "medium",
				CompactMode:  false,
				AnimationsOn: true,
			},
			Integrations: []string{"slack", "github", "jira", "datadog"},
		},
		
		Subscriptions: []Subscription{
			{
				ID:        "sub_123456",
				PlanName:  "Pro",
				Status:    "active",
				StartDate: now.Add(-365 * 24 * time.Hour),
				EndDate:   nil,
				Amount:    99.99,
				Currency:  "USD",
				Features:  []string{"advanced_analytics", "unlimited_projects", "priority_support"},
				Metadata:  map[string]string{"source": "stripe", "customer_id": "cus_abc123"},
			},
		},
		
		Permissions: []Permission{
			{
				Resource:  "projects",
				Actions:   []string{"read", "write", "delete"},
				Scope:     "organization",
				Granted:   true,
				GrantedAt: now,
				GrantedBy: "admin@example.com",
			},
			{
				Resource:  "users",
				Actions:   []string{"read"},
				Scope:     "organization",
				Granted:   true,
				GrantedAt: now,
				GrantedBy: "admin@example.com",
			},
		},
		
		SocialAccounts: []SocialAccount{
			{
				Provider:    "github",
				ExternalID:  "123456789",
				Username:    "johndoe",
				DisplayName: "John Doe",
				Email:       &phone, // reusing phone var as string pointer
				Avatar:      &phone, // reusing phone var as string pointer  
				IsVerified:  true,
				ConnectedAt: now.Add(-180 * 24 * time.Hour),
				LastSyncAt:  &now,
				Metadata:    map[string]string{"repos": "50", "followers": "120"},
			},
		},
		
		Tags:   []string{"senior", "golang", "kubernetes", "mongodb", "microservices", "devops"},
		Scores: []int{98, 95, 92, 89, 94, 97, 91, 88, 93, 96},
		
		ActivityLogs: []ActivityLog{
			{
				ID:         "log_001",
				Action:     "login",
				Resource:   "user",
				ResourceID: "1",
				Timestamp:  now.Add(-1 * time.Hour),
				IPAddress:  "192.168.1.100",
				UserAgent:  "Mozilla/5.0 Chrome/120.0",
				Result:     "success",
				Duration:   time.Millisecond * 250,
				Details: map[string]interface{}{
					"method":    "oauth",
					"provider":  "github",
					"location":  "New York, NY",
					"device":    "desktop",
				},
			},
		},
		
		CustomFields: []CustomField{
			{
				Name:        "employee_id",
				Type:        "string",
				Value:       "EMP001",
				IsRequired:  true,
				IsSearchable: true,
				Category:    "employment",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		
		Addresses: []Address{
			{
				Type:       "home",
				Street:     "123 Main St",
				City:       "New York",
				State:      "NY",
				Country:    "US",
				PostalCode: "10001",
				Latitude:   &lat,
				Longitude:  &lng,
				IsPrimary:  true,
				IsVerified: true,
			},
		},
		
		Metadata: map[string]interface{}{
			"source":           "api",
			"version":          "2.0",
			"experiment_group": "A",
			"lifetime_value":   1500.50,
			"referral_count":   5,
			"last_purchase":    now.Add(-30 * 24 * time.Hour),
		},
		
		Features: map[string]bool{
			"beta_features":     true,
			"advanced_metrics":  true,
			"custom_branding":   true,
			"api_access":        true,
			"export_data":       true,
		},
		
		Configurations: map[string]Configuration{
			"email_notifications": {
				Enabled: true,
				Settings: map[string]interface{}{
					"frequency": "daily",
					"types":     []string{"security", "billing", "product"},
				},
				UpdatedAt: now,
				UpdatedBy: "john@example.com",
			},
		},
		
		Company: &Company{
			ID:       "comp_123",
			Name:     "TechCorp Inc",
			Industry: "Technology",
			Size:     "500-1000",
			Website:  "https://techcorp.com",
			Founded:  &founded,
			Address: Address{
				Type:       "headquarters",
				Street:     "456 Tech Blvd",
				City:       "San Francisco",
				State:      "CA",
				Country:    "US",
				PostalCode: "94105",
				IsPrimary:  true,
				IsVerified: true,
			},
			Metadata: map[string]string{
				"industry_code": "5112",
				"tax_id":        "12-3456789",
			},
		},
		
		LastLoginAt:    &lastLogin,
		PasswordSetAt:  now.Add(-90 * 24 * time.Hour),
		SessionTimeout: time.Hour * 24,
	}
}

// Benchmark tests

func BenchmarkDiff_SimpleStruct(b *testing.B) {
	oldUser := createSimpleUser()
	newUser := oldUser
	newUser.Name = "Jane Doe"
	newUser.Age = 25
	newUser.Email = "jane@example.com"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_SimpleStruct_NoChanges(b *testing.B) {
	oldUser := createSimpleUser()
	newUser := oldUser
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_MediumStruct(b *testing.B) {
	oldUser := createMediumUser()
	newUser := oldUser
	newUser.Name = "Jane Smith"
	newUser.Age = 28
	newUser.Profile.Bio = "Senior Software Engineer"
	newUser.Tags = append(newUser.Tags, "kubernetes", "docker")
	newUser.Metadata["experiment"] = false
	newUser.Metadata["new_field"] = "new_value"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_MediumStruct_NoChanges(b *testing.B) {
	oldUser := createMediumUser()
	newUser := oldUser
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ComplexStruct(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	// Modify multiple fields at different nesting levels
	newUser.Name = "Jane Smith"
	newUser.Age = 28
	newUser.Profile.Bio = "Principal Software Engineer"
	newUser.Settings.Theme = "light"
	newUser.Settings.Privacy.ProfileVisible = false
	newUser.Preferences.Dashboard.Layout = "list"
	newUser.Subscriptions[0].Amount = 149.99
	newUser.Permissions = append(newUser.Permissions, Permission{
		Resource:  "billing",
		Actions:   []string{"read"},
		Scope:     "self",
		Granted:   true,
		GrantedAt: time.Now(),
		GrantedBy: "system",
	})
	newUser.Tags = append(newUser.Tags, "architect", "team-lead")
	newUser.Metadata["experiment_group"] = "B"
	newUser.Features["new_dashboard"] = true
	if newUser.Company != nil {
		newUser.Company.Size = "1000+"
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ComplexStruct_NoChanges(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ComplexStruct_ArrayChanges(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	// Focus on array changes
	newUser.Tags = append(newUser.Tags, "new-tag-1", "new-tag-2", "new-tag-3")
	newUser.Scores = append(newUser.Scores, 99, 97, 95)
	newUser.ActivityLogs = append(newUser.ActivityLogs, ActivityLog{
		ID:         "log_002",
		Action:     "logout",
		Resource:   "user",
		ResourceID: "1",
		Timestamp:  time.Now(),
		IPAddress:  "192.168.1.100",
		UserAgent:  "Mozilla/5.0 Chrome/120.0",
		Result:     "success",
		Duration:   time.Millisecond * 100,
		Details:    map[string]interface{}{"reason": "manual"},
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser, WithArrayStrategy(ArraySmart))
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ComplexStruct_WithIgnoreFields(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	// Modify fields that will be ignored
	newUser.UpdatedAt = time.Now()
	newUser.LastLoginAt = &time.Time{}
	newUser.ActivityLogs[0].Timestamp = time.Now()
	
	// Also modify fields that won't be ignored
	newUser.Name = "Jane Smith"
	newUser.Age = 28
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser, WithIgnoreFields(
			"updated_at",
			"last_login_at", 
			"activity_logs.timestamp",
		))
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

// Memory allocation benchmarks

func BenchmarkDiff_SimpleStruct_Allocs(b *testing.B) {
	oldUser := createSimpleUser()
	newUser := oldUser
	newUser.Name = "Jane Doe"
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ComplexStruct_Allocs(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	newUser.Name = "Jane Smith"
	newUser.Profile.Bio = "Principal Engineer"
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

// Comparative benchmarks for different strategies

func BenchmarkDiff_ArrayStrategies_Comparison(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	// Modify arrays significantly
	newUser.Tags = append(newUser.Tags, "new1", "new2", "new3")
	newUser.Scores = append(newUser.Scores, 99, 98, 97)
	newUser.ActivityLogs[0].Details["modified"] = true

	strategies := []struct {
		name     string
		strategy ArrayStrategy
	}{
		{"Replace", ArrayReplace},
		{"Smart", ArraySmart},
		{"Append", ArrayAppend},
		{"Merge", ArrayMerge},
	}

	for _, strategy := range strategies {
		b.Run(strategy.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				patch, err := Diff(oldUser, newUser, WithArrayStrategy(strategy.strategy))
				if err != nil {
					b.Fatal(err)
				}
				_ = patch
			}
		})
	}
}

func BenchmarkDiff_StructSize_Scaling(b *testing.B) {
	benchmarks := []struct {
		name string
		fn   func() (interface{}, interface{})
	}{
		{
			name: "Simple_5_Fields",
			fn: func() (interface{}, interface{}) {
				old := createSimpleUser()
				new := old
				new.Name = "Changed"
				new.Age = 99
				return old, new
			},
		},
		{
			name: "Medium_15_Fields", 
			fn: func() (interface{}, interface{}) {
				old := createMediumUser()
				new := old
				new.Name = "Changed"
				new.Profile.Bio = "New bio"
				new.Tags = append(new.Tags, "new")
				return old, new
			},
		},
		{
			name: "Complex_50+_Fields",
			fn: func() (interface{}, interface{}) {
				old := createComplexUser()
				new := old
				new.Name = "Changed"
				new.Profile.Bio = "New bio"
				new.Settings.Theme = "light"
				new.Tags = append(new.Tags, "new")
				return old, new
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			old, new := bm.fn()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				patch, err := Diff(old, new)
				if err != nil {
					b.Fatal(err)
				}
				_ = patch
			}
		})
	}
}

// CPU and memory profiling helpers

func BenchmarkDiff_ProfileCPU(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	
	// Make comprehensive changes
	newUser.Name = "Jane Smith"
	newUser.Age = 28
	newUser.Profile.Bio = "Principal Engineer"
	newUser.Settings.Theme = "light"
	newUser.Tags = append(newUser.Tags, "architect", "team-lead")
	newUser.Metadata["test"] = "value"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

func BenchmarkDiff_ProfileMemory(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	newUser.Name = "Changed"
	
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patch, err := Diff(oldUser, newUser)
		if err != nil {
			b.Fatal(err)
		}
		_ = patch
	}
}

// JSON marshaling benchmarks

func BenchmarkPatch_JSONMarshal_Simple(b *testing.B) {
	oldUser := createSimpleUser()
	newUser := oldUser
	newUser.Name = "Jane Doe"
	
	patch, err := Diff(oldUser, newUser)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := patch.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPatch_JSONMarshal_Complex(b *testing.B) {
	oldUser := createComplexUser()
	newUser := oldUser
	newUser.Name = "Jane Smith"
	newUser.Profile.Bio = "Principal Engineer"
	newUser.Tags = append(newUser.Tags, "architect")
	
	patch, err := Diff(oldUser, newUser)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := patch.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}