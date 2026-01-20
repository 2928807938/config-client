package service

import (
	"context"
	"strings"

	"config-client/config/domain/entity"
	domainErrors "config-client/config/domain/errors"
	"config-client/config/domain/repository"
)

// NamespaceService 命名空间领域服务
// 负责处理命名空间相关的复杂业务逻辑，包括：
// - 命名空间的创建、更新、删除等业务流程
// - 命名空间的验证和校验
// - 命名空间的状态管理
type NamespaceService struct {
	namespaceRepo repository.NamespaceRepository
	configRepo    repository.ConfigRepository
}

// NewNamespaceService 创建命名空间领域服务实例
func NewNamespaceService(
	namespaceRepo repository.NamespaceRepository,
	configRepo repository.ConfigRepository,
) *NamespaceService {
	return &NamespaceService{
		namespaceRepo: namespaceRepo,
		configRepo:    configRepo,
	}
}

// CreateNamespace 创建命名空间
// 业务规则：
// 1. 命名空间名称不能为空，且必须符合命名规范
// 2. 命名空间名称必须全局唯一
// 3. 默认状态为激活
func (s *NamespaceService) CreateNamespace(ctx context.Context, namespace *entity.Namespace) error {
	// 1. 验证命名空间有效性
	if err := s.ValidateNamespace(ctx, namespace); err != nil {
		return err
	}

	// 2. 检查命名空间是否已存在
	exists, err := s.namespaceRepo.ExistsByName(ctx, namespace.Name)
	if err != nil {
		return err
	}
	if exists {
		return domainErrors.ErrNamespaceAlreadyExists(namespace.Name)
	}

	// 3. 设置默认值
	namespace.IsActive = true // 新创建的命名空间默认激活

	if namespace.DisplayName == "" {
		namespace.DisplayName = namespace.Name
	}

	if namespace.Metadata == "" {
		namespace.Metadata = "{}"
	}

	// 4. 保存命名空间
	return s.namespaceRepo.Create(ctx, namespace)
}

// UpdateNamespace 更新命名空间
// 业务规则：
// 1. 命名空间必须存在
// 2. 命名空间名称不可修改（唯一标识）
// 3. 可以修改显示名称、描述、元数据
func (s *NamespaceService) UpdateNamespace(ctx context.Context, namespace *entity.Namespace) error {
	// 1. 检查命名空间是否存在
	existingNamespace, err := s.namespaceRepo.GetByID(ctx, namespace.ID)
	if err != nil {
		return err
	}
	if existingNamespace == nil {
		return domainErrors.ErrNamespaceNotFound(namespace.Name)
	}

	// 2. 验证命名空间有效性（仅验证可修改字段）
	if namespace.DisplayName == "" {
		return domainErrors.ErrNamespaceDisplayNameEmpty()
	}

	// 3. 使用领域实体的方法更新信息
	existingNamespace.UpdateInfo(namespace.DisplayName, namespace.Description, namespace.Metadata)

	// 4. 保存更新
	return s.namespaceRepo.Update(ctx, existingNamespace)
}

// DeleteNamespace 删除命名空间（软删除）
// 业务规则：
// 1. 命名空间必须存在
// 2. 删除前需要先停用
// 3. 确保该命名空间下没有关联的配置
func (s *NamespaceService) DeleteNamespace(ctx context.Context, id int) error {
	// 1. 检查命名空间是否存在
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if namespace == nil {
		return domainErrors.ErrNamespaceNotFound("")
	}

	// 2. 检查命名空间是否已停用
	if namespace.IsActive {
		return domainErrors.ErrNamespaceMustDeactivate(namespace.Name)
	}

	// 3. 检查是否存在关联的配置
	configCount, err := s.configRepo.CountByNamespace(ctx, id)
	if err != nil {
		return err
	}
	if configCount > 0 {
		return domainErrors.ErrNamespaceCannotDelete(namespace.Name, configCount)
	}

	// 4. 执行软删除
	return s.namespaceRepo.Delete(ctx, id)
}

// ActivateNamespace 激活命名空间
// 业务规则：
// 1. 命名空间必须存在
// 2. 验证命名空间有效性
func (s *NamespaceService) ActivateNamespace(ctx context.Context, id int) error {
	// 1. 查询命名空间
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if namespace == nil {
		return domainErrors.ErrNamespaceNotFound("")
	}

	// 2. 检查是否已经激活
	if namespace.IsActive {
		return nil // 已经激活，无需操作
	}

	// 3. 激活命名空间（使用领域实体的方法）
	namespace.Activate()

	// 4. 保存更新
	return s.namespaceRepo.Update(ctx, namespace)
}

// DeactivateNamespace 停用命名空间
// 业务规则：
// 1. 命名空间必须存在
// 2. 停用后该命名空间下的配置将不可访问
func (s *NamespaceService) DeactivateNamespace(ctx context.Context, id int) error {
	// 1. 查询命名空间
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if namespace == nil {
		return domainErrors.ErrNamespaceNotFound("")
	}

	// 2. 检查是否已经停用
	if !namespace.IsActive {
		return nil // 已经停用，无需操作
	}

	// 3. 停用命名空间（使用领域实体的方法）
	namespace.Deactivate()

	// 4. 保存更新
	return s.namespaceRepo.Update(ctx, namespace)
}

// ValidateNamespace 验证命名空间的有效性
// 验证规则：
// 1. 名称符合命名规范（字母、数字、下划线、中划线）
// 2. 名称长度限制（2-255 字符）
func (s *NamespaceService) ValidateNamespace(ctx context.Context, namespace *entity.Namespace) error {
	// 1. 验证名称
	if namespace.Name == "" {
		return domainErrors.ErrNamespaceNameEmpty()
	}

	// 2. 验证名称长度
	if len(namespace.Name) < 2 || len(namespace.Name) > 255 {
		return domainErrors.ErrNamespaceNameLengthInvalid(namespace.Name)
	}

	// 3. 验证名称格式：只允许小写字母、数字、下划线、中划线
	if !isValidNamespaceName(namespace.Name) {
		return domainErrors.ErrNamespaceNameFormatInvalid(namespace.Name)
	}

	// 4. 验证显示名称
	if namespace.DisplayName != "" && len(namespace.DisplayName) > 255 {
		return domainErrors.ErrNamespaceDisplayNameTooLong()
	}

	return nil
}

// GetActiveNamespace 获取激活的命名空间
// 业务规则：
// 1. 命名空间必须存在
// 2. 命名空间必须已激活
func (s *NamespaceService) GetActiveNamespace(ctx context.Context, name string) (*entity.Namespace, error) {
	// 1. 查询命名空间
	namespace, err := s.namespaceRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		return nil, domainErrors.ErrNamespaceNotFound(name)
	}

	// 2. 检查命名空间是否已激活
	if !namespace.IsActive {
		return nil, domainErrors.ErrNamespaceNotActive(name)
	}

	return namespace, nil
}

// ==================== 辅助函数 ====================

// isValidNamespaceName 验证命名空间名称是否符合命名规范
// 规则：只允许小写字母、数字、下划线、中划线
func isValidNamespaceName(name string) bool {
	if name == "" {
		return false
	}

	// 转换为小写并验证
	name = strings.ToLower(name)

	// 只允许小写字母、数字、下划线、中划线
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') ||
			c == '_' || c == '-') {
			return false
		}
	}

	// 不能以数字或特殊字符开头
	firstChar := rune(name[0])
	if !(firstChar >= 'a' && firstChar <= 'z') {
		return false
	}

	return true
}
