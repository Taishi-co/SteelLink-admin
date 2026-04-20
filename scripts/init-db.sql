-- SteelLink Admin Database Initialization Script
-- This script initializes the shared database for SteelLink Admin

-- Set character set
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- Admin Users Table
CREATE TABLE IF NOT EXISTS `admin_users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Username',
    `password` VARCHAR(255) NOT NULL COMMENT 'Password hash',
    `real_name` VARCHAR(50) DEFAULT '' COMMENT 'Real name',
    `email` VARCHAR(100) DEFAULT '' COMMENT 'Email',
    `phone` VARCHAR(20) DEFAULT '' COMMENT 'Phone number',
    `avatar` VARCHAR(255) DEFAULT '' COMMENT 'Avatar URL',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-disabled, 1-enabled',
    `role_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Role ID',
    `last_login_at` DATETIME DEFAULT NULL COMMENT 'Last login time',
    `last_login_ip` VARCHAR(45) DEFAULT '' COMMENT 'Last login IP',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_username` (`username`),
    INDEX `idx_status` (`status`),
    INDEX `idx_role_id` (`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin users table';

-- Admin Roles Table
CREATE TABLE IF NOT EXISTS `admin_roles` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL COMMENT 'Role name',
    `code` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Role code',
    `description` VARCHAR(255) DEFAULT '' COMMENT 'Role description',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-disabled, 1-enabled',
    `sort` INT DEFAULT 0 COMMENT 'Sort order',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_code` (`code`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin roles table';

-- Admin Menus Table
CREATE TABLE IF NOT EXISTS `admin_menus` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Parent menu ID',
    `name` VARCHAR(50) NOT NULL COMMENT 'Menu name',
    `path` VARCHAR(255) DEFAULT '' COMMENT 'Route path',
    `component` VARCHAR(255) DEFAULT '' COMMENT 'Component path',
    `icon` VARCHAR(50) DEFAULT '' COMMENT 'Icon name',
    `type` TINYINT DEFAULT 1 COMMENT 'Type: 1-menu, 2-button',
    `permission` VARCHAR(100) DEFAULT '' COMMENT 'Permission code',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-hidden, 1-visible',
    `sort` INT DEFAULT 0 COMMENT 'Sort order',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin menus table';

-- Role Menu Association Table
CREATE TABLE IF NOT EXISTS `admin_role_menus` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT 'Role ID',
    `menu_id` BIGINT UNSIGNED NOT NULL COMMENT 'Menu ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_role_menu` (`role_id`, `menu_id`),
    INDEX `idx_role_id` (`role_id`),
    INDEX `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Role menu association table';

-- Admin Operation Logs Table
CREATE TABLE IF NOT EXISTS `admin_operation_logs` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `admin_id` BIGINT UNSIGNED NOT NULL COMMENT 'Admin ID',
    `admin_name` VARCHAR(50) DEFAULT '' COMMENT 'Admin name',
    `module` VARCHAR(50) DEFAULT '' COMMENT 'Module name',
    `action` VARCHAR(50) DEFAULT '' COMMENT 'Action name',
    `method` VARCHAR(10) DEFAULT '' COMMENT 'HTTP method',
    `path` VARCHAR(255) DEFAULT '' COMMENT 'Request path',
    `ip` VARCHAR(45) DEFAULT '' COMMENT 'Request IP',
    `user_agent` VARCHAR(255) DEFAULT '' COMMENT 'User agent',
    `request_body` TEXT COMMENT 'Request body',
    `response_code` INT DEFAULT 0 COMMENT 'Response code',
    `response_msg` VARCHAR(255) DEFAULT '' COMMENT 'Response message',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_admin_id` (`admin_id`),
    INDEX `idx_module` (`module`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin operation logs table';

-- Users Table (shared with SteelLink-backend)
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Username',
    `password` VARCHAR(255) NOT NULL COMMENT 'Password hash',
    `phone` VARCHAR(20) DEFAULT '' COMMENT 'Phone number',
    `email` VARCHAR(100) DEFAULT '' COMMENT 'Email',
    `real_name` VARCHAR(50) DEFAULT '' COMMENT 'Real name',
    `avatar` VARCHAR(255) DEFAULT '' COMMENT 'Avatar URL',
    `enterprise_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Enterprise ID',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-disabled, 1-enabled',
    `last_login_at` DATETIME DEFAULT NULL COMMENT 'Last login time',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_username` (`username`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_enterprise_id` (`enterprise_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Users table';

-- Enterprises Table (shared with SteelLink-backend)
CREATE TABLE IF NOT EXISTS `enterprises` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL COMMENT 'Enterprise name',
    `short_name` VARCHAR(50) DEFAULT '' COMMENT 'Short name',
    `credit_code` VARCHAR(50) DEFAULT '' COMMENT 'Unified social credit code',
    `legal_person` VARCHAR(50) DEFAULT '' COMMENT 'Legal person name',
    `contact_person` VARCHAR(50) DEFAULT '' COMMENT 'Contact person',
    `contact_phone` VARCHAR(20) DEFAULT '' COMMENT 'Contact phone',
    `address` VARCHAR(255) DEFAULT '' COMMENT 'Address',
    `province` VARCHAR(50) DEFAULT '' COMMENT 'Province',
    `city` VARCHAR(50) DEFAULT '' COMMENT 'City',
    `district` VARCHAR(50) DEFAULT '' COMMENT 'District',
    `business_license` VARCHAR(255) DEFAULT '' COMMENT 'Business license URL',
    `status` TINYINT DEFAULT 0 COMMENT 'Status: 0-pending, 1-approved, 2-rejected',
    `audit_remark` VARCHAR(500) DEFAULT '' COMMENT 'Audit remark',
    `auditor_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Auditor ID',
    `audited_at` DATETIME DEFAULT NULL COMMENT 'Audit time',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_name` (`name`),
    INDEX `idx_credit_code` (`credit_code`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Enterprises table';

-- Products Table (shared with SteelLink-backend)
CREATE TABLE IF NOT EXISTS `products` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(200) NOT NULL COMMENT 'Product name',
    `category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Category ID',
    `enterprise_id` BIGINT UNSIGNED NOT NULL COMMENT 'Enterprise ID',
    `spec` VARCHAR(200) DEFAULT '' COMMENT 'Specification',
    `unit` VARCHAR(20) DEFAULT '' COMMENT 'Unit',
    `price` DECIMAL(12,2) DEFAULT 0 COMMENT 'Price',
    `stock` INT DEFAULT 0 COMMENT 'Stock',
    `description` TEXT COMMENT 'Description',
    `images` JSON COMMENT 'Images',
    `status` TINYINT DEFAULT 0 COMMENT 'Status: 0-pending, 1-approved, 2-rejected',
    `audit_remark` VARCHAR(500) DEFAULT '' COMMENT 'Audit remark',
    `auditor_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Auditor ID',
    `audited_at` DATETIME DEFAULT NULL COMMENT 'Audit time',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_name` (`name`),
    INDEX `idx_category_id` (`category_id`),
    INDEX `idx_enterprise_id` (`enterprise_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Products table';

-- Product Categories Table (shared with SteelLink-backend)
CREATE TABLE IF NOT EXISTS `product_categories` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `parent_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Parent category ID',
    `name` VARCHAR(50) NOT NULL COMMENT 'Category name',
    `code` VARCHAR(50) DEFAULT '' COMMENT 'Category code',
    `icon` VARCHAR(255) DEFAULT '' COMMENT 'Icon URL',
    `sort` INT DEFAULT 0 COMMENT 'Sort order',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-disabled, 1-enabled',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_code` (`code`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Product categories table';

-- Orders Table (shared with SteelLink-backend, read-only for admin)
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `order_no` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Order number',
    `buyer_id` BIGINT UNSIGNED NOT NULL COMMENT 'Buyer user ID',
    `seller_enterprise_id` BIGINT UNSIGNED NOT NULL COMMENT 'Seller enterprise ID',
    `total_amount` DECIMAL(12,2) DEFAULT 0 COMMENT 'Total amount',
    `status` TINYINT DEFAULT 0 COMMENT 'Order status',
    `remark` VARCHAR(500) DEFAULT '' COMMENT 'Remark',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` DATETIME DEFAULT NULL,
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_buyer_id` (`buyer_id`),
    INDEX `idx_seller_enterprise_id` (`seller_enterprise_id`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Orders table';

-- Insert default super admin
INSERT INTO `admin_users` (`username`, `password`, `real_name`, `status`) VALUES
('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iAt6Z5EH', 'Super Admin', 1);

-- Insert default roles
INSERT INTO `admin_roles` (`name`, `code`, `description`, `sort`) VALUES
('Super Admin', 'super_admin', 'Full system access', 1),
('Operations Admin', 'ops_admin', 'Operations management', 2),
('Auditor', 'auditor', 'Content auditing', 3);

-- Insert default menus
INSERT INTO `admin_menus` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `permission`, `sort`) VALUES
(0, 'Dashboard', '/dashboard', 'dashboard/index', 'DashboardOutlined', 1, 'dashboard', 1),
(0, 'System', '/system', '', 'SettingOutlined', 1, 'system', 2),
(2, 'Admin Management', '/system/admin', 'system/admin/index', '', 1, 'system:admin', 1),
(2, 'Role Management', '/system/role', 'system/role/index', '', 1, 'system:role', 2),
(2, 'Menu Management', '/system/menu', 'system/menu/index', '', 1, 'system:menu', 3),
(0, 'User Management', '/user', '', 'UserOutlined', 1, 'user', 3),
(5, 'User List', '/user/list', 'user/index', '', 1, 'user:list', 1),
(0, 'Enterprise Management', '/enterprise', '', 'BankOutlined', 1, 'enterprise', 4),
(6, 'Enterprise List', '/enterprise/list', 'enterprise/index', '', 1, 'enterprise:list', 1),
(0, 'Product Management', '/product', '', 'ShoppingOutlined', 1, 'product', 5),
(7, 'Product List', '/product/list', 'product/index', '', 1, 'product:list', 1),
(7, 'Category Management', '/product/category', 'product/category/index', '', 1, 'product:category', 2);

-- Associate all menus with super admin role
INSERT INTO `admin_role_menus` (`role_id`, `menu_id`)
SELECT 1, id FROM `admin_menus`;
