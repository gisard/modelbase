# ModelBase

ModelBase 是一个 Go 语言库，旨在简化数据模型层的开发工作，提供了一套通用的数据库操作接口，帮助开发者减少重复代码编写。

## 功能特性

- 预先注入类型的方式，为返回结果指定具体类型，利用 Go 泛型特性提供类型安全的数据操作
- 聚合常用数据库操作方法（增删改查、批量操作、条件查询等）
- 支持分页查询、排序和条件过滤等高级查询功能
- 支持数据库行锁操作
- **可扩展架构**，支持根据业务需求定制专用的 ModelBase 实现

## 技术依赖

- 基于 [GORM](https://gorm.io/) 框架构建
- 要求 Go 1.18 或更高版本
- 目前仅支持指针形式（`*p`）的类型注入，这是继承自 GORM 的特性

## 安装

```bash
go get github.com/gisard/modelbase
```

## 基本用法

### 创建 ModelBase 实例

```go
import (
    "github.com/gisard/modelbase"
    "gorm.io/gorm"
)

// 假设 UserModel 实现了 DataObjecter 接口
func NewUserModel(db *gorm.DB) modelbase.ModelBase[int64, *UserModel] {
    return modelbase.NewModelBase[int64, *UserModel](db)
}
```

### 获取底层 GORM 连接

为了满足更多自定义场景需求，可以通过 `GetDB` 方法获取原生 `*gorm.DB` 对象：

```go
db := userModel.GetDB(ctx)
// 使用原生 GORM API 进行自定义操作
```

## 扩展与定制

ModelBase 设计为可扩展的架构，允许开发者根据特定业务需求创建自定义的 ModelBase 实现。

### DelModelBase - 软删除支持

项目包含了一个名为 `delmodelbase` 的扩展实现，专为需要软删除功能的应用场景设计：

- 自动处理软删除字段（`is_deleted`）
- 所有查询操作都会自动过滤掉已删除的记录
- 支持创建时间和更新时间的自动管理
- 在数据更新时自动设置更新时间

#### 使用 DelModelBase

```go
import (
    "github.com/gisard/modelbase/delmodelbase"
    "gorm.io/gorm"
)

type UserModel struct {
    delmodelbase.DelObject
}

// 模型需要实现 delmodelbase.DataObjecter 接口
func NewUserModel(db *gorm.DB) delmodelbase.ModelBase[int64, *UserModel] {
    return delmodelbase.NewModelBase[int64, *UserModel](db)
}

// 使用示例
func GetUser(ctx context.Context, userID int64) (*UserModel, error) {
    // 查询会自动添加 is_deleted = 0 的过滤条件
    return userModel.Get(ctx, userID)
}

// 软删除示例
func DeleteUser(ctx context.Context, userID int64) error {
    // 实际会将记录的 is_deleted 字段设置为 1，而不是物理删除
    return userModel.Delete(ctx, userID)
}
```

### 创建自定义 ModelBase 扩展

您可以参考 `delmodelbase` 的实现方式，根据自身业务需求创建定制的 ModelBase 扩展：

1. 定义扩展的 `DataObjecter` 接口，添加所需的额外方法
2. 创建新的 `ModelBase` 接口，嵌入基础 `modelbase.ModelBase` 接口
3. 实现自定义的 modelBase 结构体，在各操作方法中添加特定的业务逻辑
4. 提供 `NewModelBase` 工厂函数创建实例

## 更多信息

详细 API 文档和更多使用示例，请查看源代码中的接口定义和测试文件。

