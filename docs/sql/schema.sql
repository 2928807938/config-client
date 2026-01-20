-- ============================================================================
-- 配置中心数据库表结构设计
-- 数据库: PostgreSQL 16
-- 编码: UTF-8
--
-- 注意：本设计不使用数据库外键约束，数据一致性由应用层维护
-- 优点：更灵活的性能优化、避免级联删除带来的风险、更容易实现分布式架构
-- ============================================================================

-- ============================================================================
-- 1. 命名空间表 (t_namespaces)
-- 用途: 隔离不同应用的配置，例如：user-service、order-service、payment-app
-- ============================================================================
CREATE TABLE t_namespaces (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 基本信息
    name VARCHAR(255) UNIQUE NOT NULL,              -- 命名空间名称，唯一标识
    display_name VARCHAR(255),                      -- 显示名称
    description TEXT,                               -- 描述信息

    -- 状态管理
    is_active BOOLEAN DEFAULT true,                 -- 是否启用
    is_deleted BOOLEAN DEFAULT false,               -- 是否删除（软删除）

    -- 审计字段
    created_by VARCHAR(100) DEFAULT 'system',       -- 创建人
    updated_by VARCHAR(100) DEFAULT 'system',       -- 更新人
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 更新时间
    deleted_at TIMESTAMP,                           -- 删除时间

    -- 扩展字段
    metadata JSONB DEFAULT '{}'::jsonb              -- 扩展元数据（JSON格式）
);

-- 索引
CREATE INDEX idx_t_namespaces_name ON t_namespaces(name) WHERE is_deleted = false;
CREATE INDEX idx_t_namespaces_active ON t_namespaces(is_active) WHERE is_deleted = false;

-- 注释
COMMENT ON TABLE t_namespaces IS '命名空间表，用于隔离不同应用的配置';
COMMENT ON COLUMN t_namespaces.name IS '命名空间名称，全���唯一，例如：user-service、order-service';
COMMENT ON COLUMN t_namespaces.display_name IS '显示名称，用于界面展示';
COMMENT ON COLUMN t_namespaces.metadata IS '扩展元数据，可存储自定义配置（JSON格式）';
COMMENT ON COLUMN t_namespaces.id IS '主键ID，自增';
COMMENT ON COLUMN t_namespaces.description IS '命名空间描述信息';
COMMENT ON COLUMN t_namespaces.is_active IS '是否启用，true-启用，false-禁用';
COMMENT ON COLUMN t_namespaces.is_deleted IS '是否删除（软删除标记），true-已删除，false-未删除';
COMMENT ON COLUMN t_namespaces.created_by IS '创建人，记录创建该记录的用户';
COMMENT ON COLUMN t_namespaces.updated_by IS '更新人，记录最后修改该记录的用户';
COMMENT ON COLUMN t_namespaces.created_at IS '创建时间，记录创建的时间戳';
COMMENT ON COLUMN t_namespaces.updated_at IS '更新时间，记录最后修改的时间戳';
COMMENT ON COLUMN t_namespaces.deleted_at IS '删除时间，软删除的时间戳';


-- ============================================================================
-- 2. 配置项表 (t_configs)
-- 用途: 存储具体的配置项，支持键值对、JSON等格式
-- ============================================================================
CREATE TABLE t_configs (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 关联信息
    namespace_id INTEGER NOT NULL,                  -- 所属命名空间ID

    -- 配置标识
    key VARCHAR(500) NOT NULL,                      -- 配置键，例如：database.host
    group_name VARCHAR(255) DEFAULT 'default',      -- 配置分组，用于逻辑分组

    -- 配置值
    value TEXT,                                     -- 配置值（文本格式）
    value_type VARCHAR(50) DEFAULT 'string',        -- 值类型：string/json/int/float/boolean/encrypted

    -- 配置哈希（用于快速比对配置内容是否变化）
    content_hash VARCHAR(32),                       -- 配置内容的MD5哈希值
    content_hash_algorithm VARCHAR(20) DEFAULT 'md5', -- 哈希算法：md5/sha256

    -- 环境隔离
    environment VARCHAR(50) DEFAULT 'default',      -- 环境：dev/test/staging/prod

    -- 版本控制
    version INTEGER DEFAULT 1,                      -- 当前版本号
    is_released BOOLEAN DEFAULT false,              -- 是否已发布

    -- 状态管理
    is_active BOOLEAN DEFAULT true,                 -- 是否启用
    is_deleted BOOLEAN DEFAULT false,               -- 是否删除（软删除）

    -- 审计字段
    created_by VARCHAR(100) DEFAULT 'system',       -- 创建人
    updated_by VARCHAR(100) DEFAULT 'system',       -- 更新人
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 更新时间
    deleted_at TIMESTAMP,                           -- 删除时间

    -- 扩展字段
    description TEXT,                               -- ���置描述
    metadata JSONB DEFAULT '{}'::jsonb,             -- 扩展元数据

    -- 唯一约束：同一命名空间下的配置键+环境必须唯一
    UNIQUE(namespace_id, key, environment)
);

-- 索引
CREATE INDEX idx_t_configs_namespace_id ON t_configs(namespace_id) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_key ON t_configs(key) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_environment ON t_configs(environment) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_group ON t_configs(group_name) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_version ON t_configs(version) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_released ON t_configs(is_released) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_hash ON t_configs(content_hash) WHERE is_deleted = false;

-- 复合索引：优化查询性能
CREATE INDEX idx_t_configs_ns_env ON t_configs(namespace_id, environment) WHERE is_deleted = false;
CREATE INDEX idx_t_configs_ns_env_ver ON t_configs(namespace_id, environment, version) WHERE is_deleted = false;

-- 注释
COMMENT ON TABLE t_configs IS '配置项表，存储应用的所有配置';
COMMENT ON COLUMN t_configs.namespace_id IS '所属命名空间ID，关联 t_namespaces 表';
COMMENT ON COLUMN t_configs.key IS '配置键，例如：database.host、redis.port';
COMMENT ON COLUMN t_configs.value IS '配置值，存储实际配置数据';
COMMENT ON COLUMN t_configs.value_type IS '值类型：string/json/int/float/boolean/encrypted';
COMMENT ON COLUMN t_configs.content_hash IS '配置内容的MD5哈希值，用于快速比对配置是否变化';
COMMENT ON COLUMN t_configs.content_hash_algorithm IS '哈希算法，默认使用MD5';
COMMENT ON COLUMN t_configs.group_name IS '配置分组，用于逻辑分类，例如：database、cache、feature';
COMMENT ON COLUMN t_configs.environment IS '环境标识：dev/test/staging/prod';
COMMENT ON COLUMN t_configs.version IS '版本号，每次修改自动递增';
COMMENT ON COLUMN t_configs.is_released IS '是否已发布到生产环境';
COMMENT ON COLUMN t_configs.id IS '主键ID，自增';
COMMENT ON COLUMN t_configs.is_active IS '是否启用，true-启用，false-禁用';
COMMENT ON COLUMN t_configs.is_deleted IS '是否删除（软删除标记），true-已删除，false-未删除';
COMMENT ON COLUMN t_configs.created_by IS '创建人，记录创建该记录的用户';
COMMENT ON COLUMN t_configs.updated_by IS '更新人，记录最后修改该记录的用户';
COMMENT ON COLUMN t_configs.created_at IS '创建时间，记录创建的时间戳';
COMMENT ON COLUMN t_configs.updated_at IS '更新时间，记录最后修改的时间戳';
COMMENT ON COLUMN t_configs.deleted_at IS '删除时间，软删除的时间戳';
COMMENT ON COLUMN t_configs.description IS '配置描述信息';
COMMENT ON COLUMN t_configs.metadata IS '扩展元数据（JSON格式）';


-- ============================================================================
-- 3. 订阅表 (t_subscriptions)
-- 用途: 记录客户端订阅信息，用于长轮询和变更推送
-- ============================================================================
CREATE TABLE t_subscriptions (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 关联信息
    namespace_id INTEGER NOT NULL,                  -- 订阅的命名空间ID

    -- 客户端信息
    client_id VARCHAR(255) NOT NULL,                -- 客户端唯一标识
    client_ip VARCHAR(50),                          -- 客户端IP地址
    client_hostname VARCHAR(255),                   -- 客户端主机名

    -- 订阅范围
    environment VARCHAR(50) DEFAULT 'default',      -- 订阅的环境

    -- 版本跟踪
    last_version INTEGER DEFAULT 0,                 -- 客户端当前版本号

    -- 配置快照哈希（用于批量比对多个配置的变化）
    config_snapshot_hash VARCHAR(32),               -- 客户端当前配置快照的MD5

    -- 状态管理
    is_active BOOLEAN DEFAULT true,                 -- 订阅是否激活

    -- 心跳检测
    last_heartbeat_at TIMESTAMP,                    -- 最后心跳时间
    heartbeat_count INTEGER DEFAULT 0,              -- 心跳次数

    -- 统计信息
    poll_count INTEGER DEFAULT 0,                   -- 长轮询次数
    change_count INTEGER DEFAULT 0,                 -- 接收到的变更次数

    -- 审计字段
    subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 订阅时间
    unsubscribed_at TIMESTAMP,                      -- 取消订阅时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 唯一约束：同一客户端在同一命名空间+环境只能有一个订阅
    UNIQUE(namespace_id, client_id, environment)
);

-- 索引
CREATE INDEX idx_t_subscriptions_namespace_id ON t_subscriptions(namespace_id) WHERE is_active = true;
CREATE INDEX idx_t_subscriptions_client_id ON t_subscriptions(client_id) WHERE is_active = true;
CREATE INDEX idx_t_subscriptions_heartbeat ON t_subscriptions(last_heartbeat_at) WHERE is_active = true;
CREATE INDEX idx_t_subscriptions_env ON t_subscriptions(environment) WHERE is_active = true;

-- 复合索引：优化查询活跃订阅
CREATE INDEX idx_t_subscriptions_active ON t_subscriptions(namespace_id, environment, is_active) WHERE is_active = true;

-- 注释
COMMENT ON TABLE t_subscriptions IS '客户端订阅表，记录长轮询订阅信息';
COMMENT ON COLUMN t_subscriptions.client_id IS '客户端唯一标识，例如：app-instance-001';
COMMENT ON COLUMN t_subscriptions.last_version IS '客户端当前配置版本号';
COMMENT ON COLUMN t_subscriptions.config_snapshot_hash IS '客户端当前配置快照的MD5，用于批量判断配置是否变化';
COMMENT ON COLUMN t_subscriptions.last_heartbeat_at IS '最后心跳时间，用于判断客户端是否存活';
COMMENT ON COLUMN t_subscriptions.poll_count IS '长轮询总次数';
COMMENT ON COLUMN t_subscriptions.change_count IS '接收到配置变更的总次数';
COMMENT ON COLUMN t_subscriptions.id IS '主键ID，自增';
COMMENT ON COLUMN t_subscriptions.namespace_id IS '订阅的命名空间ID，关联t_namespaces表';
COMMENT ON COLUMN t_subscriptions.client_ip IS '客户端IP地址';
COMMENT ON COLUMN t_subscriptions.client_hostname IS '客户端主机名';
COMMENT ON COLUMN t_subscriptions.environment IS '订阅的环境：dev/test/staging/prod';
COMMENT ON COLUMN t_subscriptions.is_active IS '订阅是否激活，true-激活，false-未激活';
COMMENT ON COLUMN t_subscriptions.heartbeat_count IS '心跳次数';
COMMENT ON COLUMN t_subscriptions.subscribed_at IS '订阅时间，首次订阅的时间戳';
COMMENT ON COLUMN t_subscriptions.unsubscribed_at IS '取消订阅时间戳';
COMMENT ON COLUMN t_subscriptions.created_at IS '记录创建时间戳';
COMMENT ON COLUMN t_subscriptions.updated_at IS '记录更新时间戳';


-- ============================================================================
-- 4. 配置变更历史表 (t_change_history)
-- 用途: 记录所有配置的变更历史，用于审计和回滚
-- ============================================================================
CREATE TABLE t_change_history (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 关联信息
    config_id INTEGER NOT NULL,                     -- 变更的配置ID
    namespace_id INTEGER NOT NULL,                  -- 所属命名空间ID

    -- 配置快照
    config_key VARCHAR(500) NOT NULL,                -- 配置键（冗余字段，方便查询）
    environment VARCHAR(50) DEFAULT 'default',      -- 环境（冗余字段）

    -- 变更信息
    operation VARCHAR(20) NOT NULL,                 -- 操作类型：CREATE/UPDATE/DELETE/ROLLBACK
    old_value TEXT,                                 -- 变更前的值
    new_value TEXT,                                 -- 变更后的值

    -- 版本信息
    old_version INTEGER,                            -- 变更前版本号
    new_version INTEGER,                            -- 变更后版本号

    -- 操作人信息
    operator VARCHAR(100) NOT NULL,                 -- 操作人
    operator_ip VARCHAR(50),                        -- 操作人IP

    -- 变更原因
    change_reason TEXT,                             -- 变更原因说明

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- 变更时间

    -- 扩展字段
    metadata JSONB DEFAULT '{}'::jsonb              -- 扩展元数据
);

-- 索引
CREATE INDEX idx_t_change_history_config_id ON t_change_history(config_id);
CREATE INDEX idx_t_change_history_namespace_id ON t_change_history(namespace_id);
CREATE INDEX idx_t_change_history_key ON t_change_history(config_key);
CREATE INDEX idx_t_change_history_operation ON t_change_history(operation);
CREATE INDEX idx_t_change_history_created_at ON t_change_history(created_at DESC);

-- 复合索引：优化查询特定配置的变更历史
CREATE INDEX idx_t_change_history_ns_key ON t_change_history(namespace_id, config_key, created_at DESC);

-- 注释
COMMENT ON TABLE t_change_history IS '配置变更历史表，记录所有配置的变更操作';
COMMENT ON COLUMN t_change_history.operation IS '操作类型：CREATE（创建）/UPDATE（更新）/DELETE（删除）/ROLLBACK（回滚）';
COMMENT ON COLUMN t_change_history.old_value IS '变更前的配置值';
COMMENT ON COLUMN t_change_history.new_value IS '变更后的配置值';
COMMENT ON COLUMN t_change_history.change_reason IS '变更原因说明，例如：切换到新数据库服务器';
COMMENT ON COLUMN t_change_history.id IS '主键ID，自增';
COMMENT ON COLUMN t_change_history.config_id IS '变更的配置ID，关联t_configs表';
COMMENT ON COLUMN t_change_history.namespace_id IS '所属命名空间ID，关联t_namespaces表';
COMMENT ON COLUMN t_change_history.config_key IS '配置键（冗余字段，方便查询）';
COMMENT ON COLUMN t_change_history.environment IS '环境（冗余字段）：dev/test/staging/prod';
COMMENT ON COLUMN t_change_history.old_version IS '变更前版本号';
COMMENT ON COLUMN t_change_history.new_version IS '变更后版本号';
COMMENT ON COLUMN t_change_history.operator IS '操作人，执行变更的用户';
COMMENT ON COLUMN t_change_history.operator_ip IS '操作人IP地址';
COMMENT ON COLUMN t_change_history.created_at IS '变更时间，记录变更的时间戳';
COMMENT ON COLUMN t_change_history.metadata IS '扩展元数据（JSON格式）';


-- ============================================================================
-- 5. 配置发布版本表 (t_release_versions)
-- 用途: 支持配置的版本发布和灰度发布
-- ============================================================================
CREATE TABLE t_release_versions (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 关联信息
    namespace_id INTEGER NOT NULL,                  -- 所属命名空间ID
    environment VARCHAR(50) NOT NULL,               -- 发布环境

    -- 版本信息
    version INTEGER NOT NULL,                       -- 版本号
    version_name VARCHAR(255) NOT NULL,             -- 版本名称，例如：v1.0.0、v1.0.1

    -- 发布范围
    config_snapshot JSONB NOT NULL,                 -- 配置快照（该版本的所有配置）
    config_count INTEGER DEFAULT 0,                 -- 包含的配置项数量

    -- 发布状态
    status VARCHAR(20) DEFAULT 'testing',            -- 状态：testing（测试中）/published（已发布）/rollback（已回滚）
    release_type VARCHAR(20) DEFAULT 'full',        -- 发布类型：full（全量）/incremental（增量）/canary（灰度）

    -- 灰度发布
    canary_rule JSONB,                             -- 灰度规则（JSON格式）
    canary_percentage INTEGER DEFAULT 0,            -- 灰度比例（0-100）

    -- 发布信息
    released_by VARCHAR(100),                       -- 发布人
    released_at TIMESTAMP,                          -- 发布时间

    -- 回滚信息
    rollback_from_version INTEGER,                  -- 从哪个版本回滚
    rollback_by VARCHAR(100),                       -- 回滚人
    rollback_at TIMESTAMP,                          -- 回滚时间
    rollback_reason TEXT,                           -- 回滚原因

    -- 审计字段
    created_by VARCHAR(100) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 唯一约束
    UNIQUE(namespace_id, environment, version)
);

-- 索引
CREATE INDEX idx_t_release_versions_ns ON t_release_versions(namespace_id, environment);
CREATE INDEX idx_t_release_versions_version ON t_release_versions(version);
CREATE INDEX idx_t_release_versions_status ON t_release_versions(status);
CREATE INDEX idx_t_release_versions_released_at ON t_release_versions(released_at DESC);

-- 注释
COMMENT ON TABLE t_release_versions IS '配置发布版本表，支持版本管理和灰度发布';
COMMENT ON COLUMN t_release_versions.version IS '版本号，单调递增';
COMMENT ON COLUMN t_release_versions.config_snapshot IS '该版本的完整配置快照（JSON格式）';
COMMENT ON COLUMN t_release_versions.status IS '状态：testing（测试中）/published（已发布）/rollback（已回滚）';
COMMENT ON COLUMN t_release_versions.release_type IS '发布类型：full（全量发布）/incremental（增量更新）/canary（灰度发布）';
COMMENT ON COLUMN t_release_versions.canary_rule IS '灰度规则，例如：按IP、按用户ID、按百分比等';
COMMENT ON COLUMN t_release_versions.canary_percentage IS '灰度比例，0-100，表示多少比例的流量使用新版本';
COMMENT ON COLUMN t_release_versions.id IS '主键ID，自增';
COMMENT ON COLUMN t_release_versions.namespace_id IS '所属命名空间ID，关联t_namespaces表';
COMMENT ON COLUMN t_release_versions.environment IS '发布环境：dev/test/staging/prod';
COMMENT ON COLUMN t_release_versions.version_name IS '版本名称，例如：v1.0.0、v1.0.1';
COMMENT ON COLUMN t_release_versions.config_count IS '包含的配置项数量';
COMMENT ON COLUMN t_release_versions.released_by IS '发布人，执行发布的用户';
COMMENT ON COLUMN t_release_versions.released_at IS '发布时间戳';
COMMENT ON COLUMN t_release_versions.rollback_from_version IS '从哪个版本回滚';
COMMENT ON COLUMN t_release_versions.rollback_by IS '回滚人，执行回滚的用户';
COMMENT ON COLUMN t_release_versions.rollback_at IS '回滚时间戳';
COMMENT ON COLUMN t_release_versions.rollback_reason IS '回滚原因说明';
COMMENT ON COLUMN t_release_versions.created_by IS '创建人，记录创建该记录的用户';
COMMENT ON COLUMN t_release_versions.created_at IS '创建时间戳';
COMMENT ON COLUMN t_release_versions.updated_at IS '更新时间戳';


-- ============================================================================
-- 6. 配置标签表 (t_config_tags)
-- 用途: 为配置打标签，支持分组和筛选（可选功能）
-- ============================================================================
CREATE TABLE t_config_tags (
    -- 主键
    id SERIAL PRIMARY KEY,

    -- 关联信息
    config_id INTEGER NOT NULL,                     -- 配置ID

    -- 标签信息
    tag_key VARCHAR(100) NOT NULL,                  -- 标签键
    tag_value VARCHAR(255) NOT NULL,                -- 标签值

    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 唯一约束
    UNIQUE(config_id, tag_key, tag_value)
);

-- 索引
CREATE INDEX idx_t_config_tags_config_id ON t_config_tags(config_id);
CREATE INDEX idx_t_config_tags_key ON t_config_tags(tag_key);
CREATE INDEX idx_t_config_tags_value ON t_config_tags(tag_value);
CREATE INDEX idx_t_config_tags_kv ON t_config_tags(tag_key, tag_value);

-- 注释
COMMENT ON TABLE t_config_tags IS '配置标签表，用于配置分组和筛选';
COMMENT ON COLUMN t_config_tags.tag_key IS '标签键，例如：sensitive、important、database';
COMMENT ON COLUMN t_config_tags.tag_value IS '标签值，例如：true、false、high-priority';
COMMENT ON COLUMN t_config_tags.id IS '主键ID，自增';
COMMENT ON COLUMN t_config_tags.config_id IS '配置ID，关联t_configs表';
COMMENT ON COLUMN t_config_tags.created_at IS '创建时间戳';


-- ============================================================================
-- 7. 系统配置表 (t_system_configs)
-- 用途: 存储配置中心的系统级配置
-- ============================================================================
CREATE TABLE t_system_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(255) UNIQUE NOT NULL,        -- 配置键
    config_value TEXT,                              -- 配置值
    description TEXT,                               -- 描述
    is_active BOOLEAN DEFAULT true,                 -- 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_t_system_configs_key ON t_system_configs(config_key);

-- 注释
COMMENT ON TABLE t_system_configs IS '系统配置表，存储配置中心的系统级配置';
COMMENT ON COLUMN t_system_configs.id IS '主键ID，自增';
COMMENT ON COLUMN t_system_configs.config_key IS '系统配置键，例如：long.polling.timeout、max.subscriptions';
COMMENT ON COLUMN t_system_configs.config_value IS '配置值';
COMMENT ON COLUMN t_system_configs.description IS '配置描述信息';
COMMENT ON COLUMN t_system_configs.is_active IS '是否启用，true-启用，false-禁用';
COMMENT ON COLUMN t_system_configs.created_at IS '创建时间戳';
COMMENT ON COLUMN t_system_configs.updated_at IS '更新时间戳';


-- ============================================================================
-- 触发器：自动更新 updated_at 字段
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要自动更新 updated_at 的表创建触发器
CREATE TRIGGER update_t_namespaces_updated_at BEFORE UPDATE ON t_namespaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_t_configs_updated_at BEFORE UPDATE ON t_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_t_subscriptions_updated_at BEFORE UPDATE ON t_subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_t_release_versions_updated_at BEFORE UPDATE ON t_release_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_t_system_configs_updated_at BEFORE UPDATE ON t_system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- 触发器：配置变更时自动记录变更历史
-- ============================================================================

-- 先创建MD5计算函数
CREATE OR REPLACE FUNCTION md5_hash(text_value TEXT)
RETURNS VARCHAR(32) AS $$
BEGIN
    IF text_value IS NULL THEN
        RETURN NULL;
    END IF;
    RETURN MD5(text_value);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 配置变更自动记录历史
CREATE OR REPLACE FUNCTION log_config_change()
RETURNS TRIGGER AS $$
DECLARE
    operation_type VARCHAR(20);
BEGIN
    -- 判断操作类型
    IF TG_OP = 'INSERT' THEN
        operation_type := 'CREATE';
        -- 自动计算MD5
        NEW.content_hash := md5_hash(NEW.value);
        INSERT INTO t_change_history (
            config_id, namespace_id, config_key, environment,
            operation, new_value, new_version, operator
        ) VALUES (
            NEW.id, NEW.namespace_id, NEW.key, NEW.environment,
            operation_type, NEW.value, NEW.version, NEW.created_by
        );
    ELSIF TG_OP = 'UPDATE' THEN
        operation_type := 'UPDATE';
        -- 自动计算MD5
        NEW.content_hash := md5_hash(NEW.value);
        -- 只有值真正变化时才记录历史
        IF OLD.value IS DISTINCT FROM NEW.value OR OLD.content_hash IS DISTINCT FROM NEW.content_hash THEN
            INSERT INTO t_change_history (
                config_id, namespace_id, config_key, environment,
                operation, old_value, new_value, old_version, new_version, operator
            ) VALUES (
                NEW.id, NEW.namespace_id, NEW.key, NEW.environment,
                operation_type, OLD.value, NEW.value, OLD.version, NEW.version, NEW.updated_by
            );
        END IF;
    ELSIF TG_OP = 'DELETE' THEN
        operation_type := 'DELETE';
        INSERT INTO t_change_history (
            config_id, namespace_id, config_key, environment,
            operation, old_value, old_version, operator
        ) VALUES (
            OLD.id, OLD.namespace_id, OLD.key, OLD.environment,
            operation_type, OLD.value, OLD.version, 'system'
        );
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 为 t_configs 表创建触发器
CREATE TRIGGER log_config_change_trigger
    AFTER INSERT OR UPDATE OR DELETE ON t_configs
    FOR EACH ROW EXECUTE FUNCTION log_config_change();


-- ============================================================================
-- 初始化数据
-- ============================================================================

-- 插入默认命名空间
INSERT INTO t_namespaces (name, display_name, description, created_by) VALUES
('default', '默认命名空间', '系统默认命名空间', 'system'),
('demo-app', '示例应用', '配置中心演示应用', 'system');

-- 插入系统配置
INSERT INTO t_system_configs (config_key, config_value, description) VALUES
('long.polling.timeout', '30', '长轮询超时时间（秒）'),
('long.polling.max.wait', '60', '长轮询最大等待时间（秒）'),
('max.subscriptions', '10000', '最大订阅数'),
('heartbeat.interval', '60', '心跳间隔时间（秒）'),
('heartbeat.timeout', '300', '心跳超时时间（秒）');

-- 插入示例配置（用于演示）
INSERT INTO t_configs (namespace_id, key, value, group_name, environment, value_type, is_released, created_by) VALUES
(2, 'database.host', 'localhost', 'database', 'default', 'string', true, 'system'),
(2, 'database.port', '5432', 'database', 'default', 'int', true, 'system'),
(2, 'database.username', 'admin', 'database', 'default', 'string', true, 'system'),
(2, 'redis.host', 'localhost', 'cache', 'default', 'string', true, 'system'),
(2, 'redis.port', '6379', 'cache', 'default', 'int', true, 'system'),
(2, 'app.debug', 'true', 'application', 'default', 'boolean', true, 'system'),
(2, 'app.log.level', 'info', 'application', 'default', 'string', true, 'system');

-- 更新示例配置的MD5哈希（因为INSERT时触发器会自动计算，这里只是确保数据一致）
UPDATE t_configs SET content_hash = MD5(value) WHERE content_hash IS NULL;


-- ============================================================================
-- 查询视图：配置总览
-- ============================================================================
CREATE OR REPLACE VIEW v_config_overview AS
SELECT
    n.name AS namespace_name,
    c.key,
    c.value,
    c.group_name,
    c.environment,
    c.value_type,
    c.version,
    c.is_released,
    c.is_active,
    c.created_by,
    c.updated_at,
    COUNT(ch.id) AS change_count
FROM t_configs c
INNER JOIN t_namespaces n ON c.namespace_id = n.id
LEFT JOIN t_change_history ch ON c.id = ch.config_id
WHERE c.is_deleted = false
GROUP BY n.name, c.id
ORDER BY n.name, c.group_name, c.key;

COMMENT ON VIEW v_config_overview IS '配置总览视图，包含配置基本信息和变更次数';


-- ============================================================================
-- 查询视图：活跃订阅统计
-- ============================================================================
CREATE OR REPLACE VIEW v_active_subscriptions AS
SELECT
    n.name AS namespace_name,
    s.environment,
    COUNT(s.id) AS total_subscriptions,
    COUNT(CASE WHEN s.last_heartbeat_at > CURRENT_TIMESTAMP - INTERVAL '5 minutes' THEN 1 END) AS active_subscriptions,
    COUNT(CASE WHEN s.last_heartbeat_at <= CURRENT_TIMESTAMP - INTERVAL '5 minutes' OR s.last_heartbeat_at IS NULL THEN 1 END) AS inactive_subscriptions,
    SUM(s.poll_count) AS total_polls,
    SUM(s.change_count) AS total_changes
FROM t_subscriptions s
INNER JOIN t_namespaces n ON s.namespace_id = n.id
WHERE s.is_active = true
GROUP BY n.name, s.environment;

COMMENT ON VIEW v_active_subscriptions IS '活跃订阅统计视图，用于监控订阅状态';


-- ============================================================================
-- 完成脚本
-- ============================================================================
-- 执行完成后，可以通过以下查询验证表结构
-- SELECT * FROM t_namespaces;
-- SELECT * FROM v_config_overview;
-- SELECT * FROM v_active_t_subscriptions;
