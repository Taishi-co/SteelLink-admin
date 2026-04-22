package store

import "time"

type Enterprise struct {
	ID               uint64     `gorm:"column:id"`
	EnterpriseCode   string     `gorm:"column:enterprise_code"`
	Name             string     `gorm:"column:name"`
	CreditCode       string     `gorm:"column:credit_code"`
	EnterpriseRole   string     `gorm:"column:enterprise_role"`
	Remark           string     `gorm:"column:remark"`
	Status           int        `gorm:"column:status"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at"`
}

func (Enterprise) TableName() string {
	return "enterprises"
}

type Product struct {
	ID            uint64     `gorm:"column:id"`
	Name          string     `gorm:"column:name"`
	ProductNo     string     `gorm:"column:product_no"`
	Category      string     `gorm:"column:category"`
	SubCategory   string     `gorm:"column:sub_category"`
	EnterpriseID  uint64     `gorm:"column:enterprise_id"`
	Spec          string     `gorm:"column:spec"`
	UnitPrice     float64    `gorm:"column:unit_price"`
	Quantity      float64    `gorm:"column:quantity"`
	AvailableQty  float64    `gorm:"column:available_qty"`
	Status        int        `gorm:"column:status"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at"`
}

func (Product) TableName() string {
	return "products"
}
