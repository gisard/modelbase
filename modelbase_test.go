package modelbase

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	ast "github.com/stretchr/testify/assert"
	st "github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBObject struct {
	ID   int64  `gorm:"type:int(11);primaryKey;autoIncrement"`
	Name string `gorm:"type:varchar(50);uniqueIndex;NOT NULL"`
	Age  int    `gorm:"type:int(11);NOT NULL;default:18"`
}

func (d *DBObject) TableName() string {
	return "user"
}

func (d *DBObject) GetID() int64 {
	return d.ID
}

type model struct {
	ModelBase[int64, *DBObject]
}

type modelTestSuite struct {
	st.Suite

	ctx     context.Context
	assert  *ast.Assertions
	sqlMock sqlmock.Sqlmock
	model   *model
}

func (m *modelTestSuite) SetupTest() {
	m.assert = ast.New(m.T())
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	m.assert.Nil(err)
	m.sqlMock = mock
	gormDB, err := gorm.Open(
		mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true, DontSupportForShareClause: true}),
		&gorm.Config{SkipDefaultTransaction: true},
	)
	m.model = &model{ModelBase: NewModelBase[int64, *DBObject](gormDB)}
	m.assert.Nil(err)

	m.ctx = context.Background()
}

func (m *modelTestSuite) TestInsertBatch() {
	userDOs := []*DBObject{
		{Name: "John", Age: 18},
		{Name: "Mary", Age: 20},
	}
	m.sqlMock.ExpectExec("INSERT INTO `user` (`name`,`age`) VALUES (?,?),(?,?)").
		WithArgs("John", 18, "Mary", 20).WillReturnResult(sqlmock.NewResult(1, 2))

	err := m.model.InsertBatch(m.ctx, userDOs...)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func (m *modelTestSuite) TestUpsert() {
	userDO := &DBObject{Name: "John", Age: 18}
	m.sqlMock.ExpectExec("INSERT INTO `user` (`name`,`age`) VALUES (?,?)").
		WithArgs("John", 18).WillReturnResult(sqlmock.NewResult(1, 1))

	err := m.model.Upsert(m.ctx, userDO)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func (m *modelTestSuite) TestGet() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `user`.`id` = ? ORDER BY `user`.`id` LIMIT 1").
		WithArgs(userID).
		WillReturnRows(queryRows)

	actualTaskDO, err := m.model.Get(m.ctx, userID)
	expectedUserDO := DBObject{
		ID:   5,
		Name: "John",
		Age:  18,
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(&expectedUserDO, actualTaskDO)
}

func (m *modelTestSuite) TestGetBy() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `id` = ? ORDER BY `user`.`id` LIMIT 1").
		WithArgs(userID).
		WillReturnRows(queryRows)

	actualTaskDO, err := m.model.GetBy(m.ctx, "`id` = ?", userID)
	expectedUserDO := DBObject{
		ID:   5,
		Name: "John",
		Age:  18,
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(&expectedUserDO, actualTaskDO)
}

func (m *modelTestSuite) TestUpdate() {
	userDO := &DBObject{ID: 4, Name: "John", Age: 18}
	m.sqlMock.ExpectExec("UPDATE `user` SET `name`=?,`age`=? WHERE `id` = ?").
		WithArgs(userDO.Name, userDO.Age, userDO.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := m.model.Update(m.ctx, userDO)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func (m *modelTestSuite) TestUpdateBatch() {
	userDO := &DBObject{ID: 4, Name: "Tom", Age: 18}
	m.sqlMock.ExpectExec("UPDATE `user` SET `name`=? WHERE `id` = ?").
		WithArgs(userDO.Name, userDO.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := m.model.UpdateBatch(m.ctx, map[string]any{"name": "Tom"}, "`id` = ?", 4)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func (m *modelTestSuite) TestList() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` ").
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx)
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListWithWhere() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `id` = ?").
		WithArgs(userID).
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx, WhereOpt("`id` = ?", userID))
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListWithPage() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` LIMIT 10 OFFSET 10").
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx, PageOpt(2, 10))
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListWithOffset() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` LIMIT 10 OFFSET 10").
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx, OffsetOpt(10, 10))
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListWithSort() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` ORDER BY `id` DESC").
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx, SortOpt("id", DESC))
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListWithWhereSortPage() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `id` = ? ORDER BY `id` DESC LIMIT 10 OFFSET 10").
		WithArgs(userID).
		WillReturnRows(queryRows)

	actualList, err := m.model.List(m.ctx, WhereOpt("`id` = ?", userID), SortOpt("id", DESC), PageOpt(2, 10))
	expectedUserDOs := []*DBObject{
		{
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserDOs, actualList)
}

func (m *modelTestSuite) TestListMap() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` ").
		WillReturnRows(queryRows)

	actualListMap, err := m.model.ListMap(m.ctx, "")
	expectedUserMap := map[int64]*DBObject{
		5: {
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserMap, actualListMap)
}

func (m *modelTestSuite) TestListMapByIDs() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `id` IN (?)").
		WithArgs(userID).
		WillReturnRows(queryRows)

	actualTaskDO, err := m.model.ListMapByIDs(m.ctx, []int64{userID})
	expectedUserMap := map[int64]*DBObject{
		5: {
			ID:   5,
			Name: "John",
			Age:  18,
		},
	}
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(expectedUserMap, actualTaskDO)
}

func (m *modelTestSuite) TestExist() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"id", "name", "age"}).
		AddRow(userID, "John", 18)
	m.sqlMock.ExpectQuery("SELECT * FROM `user` WHERE `id` = ? ORDER BY `user`.`id` LIMIT 1").
		WithArgs(userID).
		WillReturnRows(queryRows)

	exist, err := m.model.Exist(m.ctx, "`id` = ?", userID)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(true, exist)
}

func (m *modelTestSuite) TestCount() {
	var userID int64 = 5
	queryRows := sqlmock.NewRows([]string{"count(*)"}).
		AddRow(1)
	m.sqlMock.ExpectQuery("SELECT count(*) FROM `user` WHERE `id` = ?").
		WithArgs(userID).
		WillReturnRows(queryRows)

	exist, err := m.model.Count(m.ctx, "`id` = ?", userID)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
	m.assert.Equal(int64(1), exist)
}

func (m *modelTestSuite) TestDelete() {
	userDO := &DBObject{ID: 4, Name: "John", Age: 18}
	m.sqlMock.ExpectExec("DELETE FROM `user` WHERE `user`.`id` = ?").
		WithArgs(userDO.ID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := m.model.Delete(m.ctx, userDO)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func (m *modelTestSuite) TestDeleteBatch() {
	var userID int64 = 5
	m.sqlMock.ExpectExec("DELETE FROM `user` WHERE `id` = (?)").
		WithArgs(userID).WillReturnResult(sqlmock.NewResult(0, 1))

	err := m.model.DeleteBatch(m.ctx, "`id` = ?", userID)
	m.assert.Nil(err)
	m.assert.Nil(m.sqlMock.ExpectationsWereMet())
}

func TestDataSuite(t *testing.T) {
	st.Run(t, new(modelTestSuite))
}
