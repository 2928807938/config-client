package service

import (
	"context"

	"config-client/api/config-api/converter"
	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/dto/vo"
	domainRepo "config-client/config/domain/repository"
	domainService "config-client/config/domain/service"
)

// NamespaceAppService 命名空间应用服务
// 负责协调领域服务和数据转换，不包含业务逻辑和异常处理
// 异常由领域服务捕获并向上传递，最终由统一异常处理器处理
type NamespaceAppService struct {
	namespaceDomainService *domainService.NamespaceService
	converter              *converter.NamespaceConverter
	namespaceRepo          domainRepo.NamespaceRepository
}

// NewNamespaceAppService 创建命名空间应用服务实例
func NewNamespaceAppService(
	namespaceDomainService *domainService.NamespaceService,
	converter *converter.NamespaceConverter,
	namespaceRepo domainRepo.NamespaceRepository,
) *NamespaceAppService {
	return &NamespaceAppService{
		namespaceDomainService: namespaceDomainService,
		converter:              converter,
		namespaceRepo:          namespaceRepo,
	}
}

// CreateNamespace 创建命名空间
func (s *NamespaceAppService) CreateNamespace(ctx context.Context, req *request.CreateNamespaceRequest) (*vo.NamespaceVO, error) {
	// 1. 将请求DTO转换为领域实体
	namespace := s.converter.ToEntity(req)

	// 2. 调用领域服务创建命名空间（错误直接向上传递）
	if err := s.namespaceDomainService.CreateNamespace(ctx, namespace); err != nil {
		return nil, err
	}

	// 3. 将领域实体转换为VO返回
	return s.converter.ToVO(namespace), nil
}

// UpdateNamespace 更新命名空间
func (s *NamespaceAppService) UpdateNamespace(ctx context.Context, id int, req *request.UpdateNamespaceRequest) (*vo.NamespaceVO, error) {
	// 1. 先查询现有命名空间以获取完整信息
	existingNamespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 使用转换器更新领域实体
	s.converter.UpdateEntityFromRequest(existingNamespace, req)

	// 3. 调用领域服务更新命名空间（错误直接向上传递）
	if err := s.namespaceDomainService.UpdateNamespace(ctx, existingNamespace); err != nil {
		return nil, err
	}

	// 4. 重新查询最新命名空间并返回
	updatedNamespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(updatedNamespace), nil
}

// DeleteNamespace 删除命名空间（软删除）
func (s *NamespaceAppService) DeleteNamespace(ctx context.Context, id int) error {
	// 直接调用领域服务删除命名空间（错误直接向上传递）
	return s.namespaceDomainService.DeleteNamespace(ctx, id)
}

// ActivateNamespace 激活命名空间
func (s *NamespaceAppService) ActivateNamespace(ctx context.Context, id int) (*vo.NamespaceVO, error) {
	// 1. 调用领域服务激活命名空间（错误直接向上传递）
	if err := s.namespaceDomainService.ActivateNamespace(ctx, id); err != nil {
		return nil, err
	}

	// 2. 查询最新命名空间并返回
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(namespace), nil
}

// DeactivateNamespace 停用命名空间
func (s *NamespaceAppService) DeactivateNamespace(ctx context.Context, id int) (*vo.NamespaceVO, error) {
	// 1. 调用领域服务停用命名空间（错误直接向上传递）
	if err := s.namespaceDomainService.DeactivateNamespace(ctx, id); err != nil {
		return nil, err
	}

	// 2. 查询最新命名空间并返回
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.converter.ToVO(namespace), nil
}

// QueryNamespaces 分页查询命名空间
func (s *NamespaceAppService) QueryNamespaces(ctx context.Context, req *request.QueryNamespaceRequest) (*vo.NamespaceListVO, error) {
	// 1. 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 2. 将 DTO 转换为仓储层查询参数
	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	params := &domainRepo.NamespaceQueryParams{
		Name:     name,
		IsActive: req.IsActive,
		Page:     req.Page,
		Size:     req.PageSize,
	}

	// 3. 调用仓储层查询命名空间（错误直接向上传递）
	pageResult, err := s.namespaceRepo.Query(ctx, params)
	if err != nil {
		return nil, err
	}

	// 4. 转换为VO返回
	return &vo.NamespaceListVO{
		Total:      pageResult.Total,
		Page:       pageResult.Page,
		PageSize:   pageResult.Size,
		Namespaces: s.converter.ToVOList(pageResult.Items),
	}, nil
}

// GetNamespaceByID 根据ID获取命名空间
func (s *NamespaceAppService) GetNamespaceByID(ctx context.Context, id int) (*vo.NamespaceVO, error) {
	// 1. 调用仓储层获取命名空间（错误直接向上传递）
	namespace, err := s.namespaceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVO(namespace), nil
}

// GetNamespaceByName 根据名称获取命名空间
func (s *NamespaceAppService) GetNamespaceByName(ctx context.Context, name string) (*vo.NamespaceVO, error) {
	// 1. 调用仓储层查询命名空间（错误直接向上传递）
	namespace, err := s.namespaceRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVO(namespace), nil
}

// GetActiveNamespace 获取激活的命名空间
func (s *NamespaceAppService) GetActiveNamespace(ctx context.Context, name string) (*vo.NamespaceVO, error) {
	// 1. 调用领域服务获取激活的命名空间（错误直接向上传递）
	namespace, err := s.namespaceDomainService.GetActiveNamespace(ctx, name)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVO(namespace), nil
}

// ListAllNamespaces 获取所有命名空间（不分页）
func (s *NamespaceAppService) ListAllNamespaces(ctx context.Context) ([]*vo.NamespaceVO, error) {
	// 1. 调用仓储层查询所有命名空间（错误直接向上传递）
	namespaces, err := s.namespaceRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVOList(namespaces), nil
}

// ListActiveNamespaces 获取所有激活的命名空间（不分页）
func (s *NamespaceAppService) ListActiveNamespaces(ctx context.Context) ([]*vo.NamespaceVO, error) {
	// 1. 调用仓储层查询所有激活的命名空间（错误直接向上传递）
	namespaces, err := s.namespaceRepo.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	// 2. 转换为VO返回
	return s.converter.ToVOList(namespaces), nil
}
