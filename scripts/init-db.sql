-- SteelLink Admin Database Initialization Script
-- This script initializes the shared database for SteelLink Admin

-- Set character set
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- ============================================================================
-- Standard Fields Convention:
-- id            - BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
-- created_by    - BIGINT UNSIGNED, creator ID (0 = system)
-- created_at    - DATETIME, creation timestamp
-- updated_by    - BIGINT UNSIGNED, last modifier ID (0 = system)
-- updated_at    - DATETIME, last modification timestamp
-- deleted_at    - DATETIME, soft delete (NULL = not deleted)
-- ============================================================================

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
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_username` (`username`),
    INDEX `idx_status` (`status`),
    INDEX `idx_role_id` (`role_id`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin users table';

-- Admin Roles Table
CREATE TABLE IF NOT EXISTS `admin_roles` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(50) NOT NULL COMMENT 'Role name',
    `code` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Role code',
    `description` VARCHAR(255) DEFAULT '' COMMENT 'Role description',
    `status` TINYINT DEFAULT 1 COMMENT 'Status: 0-disabled, 1-enabled',
    `sort` INT DEFAULT 0 COMMENT 'Sort order',
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_code` (`code`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
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
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Admin menus table';

-- Role Menu Association Table
CREATE TABLE IF NOT EXISTS `admin_role_menus` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `role_id` BIGINT UNSIGNED NOT NULL COMMENT 'Role ID',
    `menu_id` BIGINT UNSIGNED NOT NULL COMMENT 'Menu ID',
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    UNIQUE KEY `uk_role_menu` (`role_id`, `menu_id`),
    INDEX `idx_role_id` (`role_id`),
    INDEX `idx_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Role menu association table';

-- Admin Operation Logs Table (no soft delete, logs are immutable)
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
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
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
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_username` (`username`),
    INDEX `idx_phone` (`phone`),
    INDEX `idx_enterprise_id` (`enterprise_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
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
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_name` (`name`),
    INDEX `idx_credit_code` (`credit_code`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Enterprises table';

-- Products Table (shared with SteelLink-backend)
CREATE TABLE IF NOT EXISTS `products` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(200) NOT NULL COMMENT 'Product name',
    `product_no` VARCHAR(50) DEFAULT '' COMMENT 'Product number',
    `category_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Category ID',
    `enterprise_id` BIGINT UNSIGNED NOT NULL COMMENT 'Enterprise ID',
    `category` VARCHAR(100) DEFAULT '' COMMENT 'Main category',
    `sub_category` VARCHAR(100) DEFAULT '' COMMENT 'Sub category',
    `steel_mill` VARCHAR(100) DEFAULT '' COMMENT 'Steel mill',
    `goods_type` VARCHAR(20) DEFAULT 'spot' COMMENT 'Goods type: spot/futures',
    `warehouse_service_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Warehouse service ID',
    `warehouse_enterprise_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Warehouse enterprise ID',
    `spec` JSON COMMENT 'Specification',
    `warehouse_info` JSON COMMENT 'Warehouse snapshot',
    `bundle_no` VARCHAR(100) DEFAULT '' COMMENT 'Bundle number',
    `unit` VARCHAR(20) DEFAULT '' COMMENT 'Unit',
    `unit_price` DECIMAL(12,2) DEFAULT 0 COMMENT 'Unit price',
    `quantity` DECIMAL(12,2) DEFAULT 0 COMMENT 'Quantity',
    `min_order_qty` DECIMAL(12,2) DEFAULT 0 COMMENT 'Minimum order quantity',
    `delivery_days` INT DEFAULT 0 COMMENT 'Delivery days',
    `quote_validity` INT DEFAULT 0 COMMENT 'Quote validity hours',
    `description` TEXT COMMENT 'Description',
    `images` JSON COMMENT 'Images',
    `status` TINYINT DEFAULT 0 COMMENT 'Status: 0-draft, 1-pending, 2-on sale, 3-off sale, 4-rejected',
    `audit_remark` VARCHAR(500) DEFAULT '' COMMENT 'Audit remark',
    `auditor_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Auditor ID',
    `approved_at` DATETIME DEFAULT NULL COMMENT 'Approved time',
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_name` (`name`),
    INDEX `idx_category_id` (`category_id`),
    INDEX `idx_enterprise_id` (`enterprise_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
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
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_parent_id` (`parent_id`),
    INDEX `idx_code` (`code`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Product categories table';

-- Orders Table (shared with SteelLink-backend, read-only for admin)
CREATE TABLE IF NOT EXISTS `orders` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `order_no` VARCHAR(50) NOT NULL UNIQUE COMMENT 'Order number',
    `buyer_id` BIGINT UNSIGNED NOT NULL COMMENT 'Buyer enterprise ID',
    `buyer_name` VARCHAR(200) DEFAULT '' COMMENT 'Buyer name',
    `seller_id` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Seller enterprise ID',
    `seller_enterprise_id` BIGINT UNSIGNED NOT NULL COMMENT 'Seller enterprise ID',
    `seller_name` VARCHAR(200) DEFAULT '' COMMENT 'Seller name',
    `total_quantity` DECIMAL(12,2) DEFAULT 0 COMMENT 'Total quantity',
    `total_amount` DECIMAL(12,2) DEFAULT 0 COMMENT 'Total amount',
    `tax_total_amount` DECIMAL(12,2) DEFAULT 0 COMMENT 'Tax total amount',
    `currency` VARCHAR(10) DEFAULT 'CNY' COMMENT 'Currency',
    `status` TINYINT DEFAULT 0 COMMENT 'Order status',
    `delivery_voucher` VARCHAR(500) DEFAULT '' COMMENT 'Delivery voucher URL',
    `receipt_voucher` VARCHAR(500) DEFAULT '' COMMENT 'Receipt voucher URL',
    `receipt_at` DATETIME DEFAULT NULL COMMENT 'Receipt time',
    `expires_at` DATETIME DEFAULT NULL COMMENT 'Expiration time',
    `remark` JSON COMMENT 'Remark',
    `created_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Creator ID',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation time',
    `updated_by` BIGINT UNSIGNED DEFAULT 0 COMMENT 'Last modifier ID',
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last modification time',
    `deleted_at` DATETIME DEFAULT NULL COMMENT 'Soft delete time',
    INDEX `idx_order_no` (`order_no`),
    INDEX `idx_buyer_id` (`buyer_id`),
    INDEX `idx_seller_enterprise_id` (`seller_enterprise_id`),
    INDEX `idx_status` (`status`),
    INDEX `idx_created_at` (`created_at`),
    INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Orders table';

-- ============================================================================
-- Initial Data
-- ============================================================================

-- Insert default super admin (password: admin123, need to hash in production)
INSERT INTO `admin_users` (`username`, `password`, `real_name`, `status`, `created_by`) VALUES
('admin', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iAt6Z5EH', 'Super Admin', 1, 0);

-- Insert default roles
INSERT INTO `admin_roles` (`name`, `code`, `description`, `sort`, `created_by`) VALUES
('Super Admin', 'super_admin', 'Full system access', 1, 0),
('Operations Admin', 'ops_admin', 'Operations management', 2, 0),
('Auditor', 'auditor', 'Content auditing', 3, 0);

-- Insert default menus
INSERT INTO `admin_menus` (`parent_id`, `name`, `path`, `component`, `icon`, `type`, `permission`, `sort`, `created_by`) VALUES
(0, 'Dashboard', '/dashboard', 'dashboard/index', 'DashboardOutlined', 1, 'dashboard', 1, 0),
(0, 'System', '/system', '', 'SettingOutlined', 1, 'system', 2, 0),
(2, 'Admin Management', '/system/admin', 'system/admin/index', '', 1, 'system:admin', 1, 0),
(2, 'Role Management', '/system/role', 'system/role/index', '', 1, 'system:role', 2, 0),
(2, 'Menu Management', '/system/menu', 'system/menu/index', '', 1, 'system:menu', 3, 0),
(0, 'Business', '/business', '', 'SolutionOutlined', 1, 'business', 3, 0),
(6, 'User Management', '/user', 'user/index', 'UserOutlined', 1, 'business:user', 1, 0),
(6, 'Enterprise Audit', '/enterprise', 'enterprise/index', 'BankOutlined', 1, 'business:enterprise', 2, 0),
(6, 'Service Permission Audit', '/enterprise/services', 'enterprise/service-applications', 'AuditOutlined', 1, 'business:enterprise-service', 3, 0),
(6, 'Warehouse Resource Audit', '/service-offers/warehouse', 'service-offer/warehouse', 'ContainerOutlined', 1, 'service:warehouse:audit', 4, 0),
(6, 'Logistics Resource Audit', '/service-offers/logistics', 'service-offer/logistics', 'TruckOutlined', 1, 'service:logistics:audit', 5, 0),
(6, 'Finance Service Audit', '/service-offers/finance', 'service-offer/finance', 'BankOutlined', 1, 'service:finance:audit', 6, 0),
(0, 'Product Center', '/product-center', '', 'ShoppingOutlined', 1, 'product', 4, 0),
(13, 'Product Management', '/product', 'product/index', 'AppstoreOutlined', 1, 'product:manage', 1, 0),
(13, 'Order Management', '/product/order', 'order/index', 'ShoppingOutlined', 1, 'order:manage', 2, 0),
(13, 'Category Management', '/product/category', 'product/category/index', 'AppstoreOutlined', 1, 'product:category', 3, 0),
(0, 'AI Procurement', '/ai-procurement', '', 'ThunderboltOutlined', 1, 'ai-procurement', 5, 0),
(17, 'Procurement Config', '/ai-procurement/config', 'ai-procurement/config/index', 'SettingOutlined', 1, 'ai-procurement:config', 1, 0),
(0, 'Operation Logs', '/log', 'log/index', 'FileTextOutlined', 1, 'log', 6, 0),
(0, 'Notifications', '/notification', 'notification/index', 'MailOutlined', 1, 'notification', 7, 0);

-- Associate all menus with super admin role
INSERT INTO `admin_role_menus` (`role_id`, `menu_id`, `created_by`)
SELECT 1, id, 0 FROM `admin_menus`;

-- Associate default business menus with operations role
INSERT INTO `admin_role_menus` (`role_id`, `menu_id`, `created_by`)
SELECT 2, id, 0 FROM `admin_menus`
WHERE `path` IN (
    '/dashboard',
    '/user',
    '/enterprise',
    '/enterprise/services',
    '/service-offers/warehouse',
    '/service-offers/logistics',
    '/service-offers/finance',
    '/product',
    '/product/order',
    '/product/category',
    '/log',
    '/notification'
);

-- Associate default audit menus with auditor role
INSERT INTO `admin_role_menus` (`role_id`, `menu_id`, `created_by`)
SELECT 3, id, 0 FROM `admin_menus`
WHERE `path` IN (
    '/dashboard',
    '/user',
    '/enterprise',
    '/enterprise/services',
    '/service-offers/warehouse',
    '/service-offers/logistics',
    '/service-offers/finance',
    '/product',
    '/product/order',
    '/product/category',
    '/log'
);
