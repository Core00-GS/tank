package rest

import (
	"fmt"
	"github.com/eyebluecn/tank/code/core"
	"github.com/eyebluecn/tank/code/tool/builder"
	"github.com/eyebluecn/tank/code/tool/result"
	"gorm.io/gorm"

	"github.com/eyebluecn/tank/code/tool/uuid"
	"time"
)

type SpaceDao struct {
	BaseDao
}

// find by uuid. if not found return nil.
func (this *SpaceDao) FindByUuid(uuid string) *Space {
	var entity = &Space{}
	db := core.CONTEXT.GetDB().Where("uuid = ?", uuid).First(entity)
	if db.Error != nil {
		if db.Error.Error() == result.DB_ERROR_NOT_FOUND {
			return nil
		} else {
			panic(db.Error)
		}
	}
	return entity
}

// find by uuid. if not found panic NotFound error
func (this *SpaceDao) CheckByUuid(uuid string) *Space {
	entity := this.FindByUuid(uuid)
	if entity == nil {
		panic(result.NotFound("not found record with uuid = %s", uuid))
	}
	return entity
}

func (this *SpaceDao) CountByName(name string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&Space{}).
		Where("name = ?", name).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

func (this *SpaceDao) CountByUserUuid(userUuid string) int {
	var count int64
	db := core.CONTEXT.GetDB().
		Model(&Space{}).
		Where("user_uuid = ?", userUuid).
		Count(&count)
	this.PanicError(db.Error)
	return int(count)
}

// TODO:
func (this *SpaceDao) SelfPage(page int, pageSize int, userUuid string, spaceType string, sortArray []builder.OrderPair) *Pager {

	countSqlTemplate := fmt.Sprintf("SELECT COUNT(*) FROM `%sspace` WHERE uuid IN (SELECT space_uuid FROM `%sspace_member` WHERE user_uuid = ?) AND type = ?", core.TABLE_PREFIX, core.TABLE_PREFIX)
	if spaceType == SPACE_TYPE_PRIVATE {
		countSqlTemplate = fmt.Sprintf("SELECT COUNT(*) FROM `%sspace` WHERE user_uuid = ? AND type = ?", core.TABLE_PREFIX)
	}
	var count int
	core.CONTEXT.GetDB().Raw(countSqlTemplate, userUuid, spaceType).Scan(&count)

	orderByString := this.GetSortString(sortArray)
	if orderByString == "" {
		orderByString = "uuid"
	}
	querySqlTemplate := fmt.Sprintf("SELECT * FROM `%sspace` WHERE uuid IN (SELECT space_uuid FROM `%sspace_member` WHERE user_uuid = ?) AND type = ? ORDER BY ? LIMIT ?,?", core.TABLE_PREFIX, core.TABLE_PREFIX)
	if spaceType == SPACE_TYPE_PRIVATE {
		querySqlTemplate = fmt.Sprintf("SELECT * FROM `%sspace` WHERE user_uuid = ? AND type = ? ORDER BY ? LIMIT ?,?", core.TABLE_PREFIX)
	}
	var spaces []*Space
	core.CONTEXT.GetDB().Raw(querySqlTemplate, userUuid, spaceType, orderByString, page*pageSize, pageSize).Scan(&spaces)

	pager := NewPager(page, pageSize, count, spaces)

	return pager

}

func (this *SpaceDao) Page(page int, pageSize int, spaceType string, sortArray []builder.OrderPair) *Pager {
	count, spaces := this.PlainPage(page, pageSize, spaceType, sortArray)
	pager := NewPager(page, pageSize, count, spaces)

	return pager
}

func (this *SpaceDao) PlainPage(page int, pageSize int, spaceType string, sortArray []builder.OrderPair) (int, []*Space) {

	var wp = &builder.WherePair{Query: "type = ?", Args: []interface{}{spaceType}}

	var conditionDB *gorm.DB
	conditionDB = core.CONTEXT.GetDB().Model(&Space{}).Where(wp.Query, wp.Args...)

	var count int64 = 0
	db := conditionDB.Count(&count)
	this.PanicError(db.Error)

	var spaces []*Space
	db = conditionDB.Order(this.GetSortString(sortArray)).Offset(page * pageSize).Limit(pageSize).Find(&spaces)
	this.PanicError(db.Error)

	return int(count), spaces
}

func (this *SpaceDao) Create(space *Space) *Space {

	timeUUID, _ := uuid.NewV4()
	space.Uuid = string(timeUUID.String())
	space.CreateTime = time.Now()
	space.UpdateTime = time.Now()
	space.Sort = time.Now().UnixNano() / 1e6
	db := core.CONTEXT.GetDB().Create(space)
	this.PanicError(db.Error)

	return space
}

func (this *SpaceDao) Save(space *Space) *Space {

	space.UpdateTime = time.Now()
	db := core.CONTEXT.GetDB().Save(space)
	this.PanicError(db.Error)

	return space
}

func (this *SpaceDao) Delete(space *Space) {

	db := core.CONTEXT.GetDB().Delete(&space)
	this.PanicError(db.Error)

}

// System cleanup.
func (this *SpaceDao) Cleanup() {
	this.logger.Info("[SpaceDao] clean up. Delete all Space")
	db := core.CONTEXT.GetDB().Where("uuid is not null").Delete(Space{})
	this.PanicError(db.Error)
}