package one

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	"jh_game_service/internal/logic/frontend/game/platform"
)

// syncMockDriver 复用 GoFrame MySQL 方言，底层连接替换为 sqlmock。
type syncMockDriver struct{ *mysql.Driver }

var (
	syncOnce sync.Once
	syncDB   *sql.DB
	syncMock sqlmock.Sqlmock
	syncMu   sync.Mutex
	syncSQL  []string
)

func (d *syncMockDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	return &syncMockDriver{Driver: &mysql.Driver{Core: core}}, nil
}

func (d *syncMockDriver) Open(config *gdb.ConfigNode) (*sql.DB, error) { return syncDB, nil }

// TableFields 返回 game_list 固定表结构。
func (d *syncMockDriver) TableFields(ctx context.Context, table string, schema ...string) (map[string]*gdb.TableField, error) {
	names := []string{"id", "site_id", "type", "platform", "name", "alias_name", "style", "img", "img1", "device", "vendor", "code", "language", "hot", "recommend", "like_num", "is_new", "sort", "status", "available", "created_at", "updated_at", "game_id", "cover_url"}
	m := make(map[string]*gdb.TableField, len(names))
	for i, n := range names {
		m[n] = &gdb.TableField{Index: i, Name: n}
	}
	return m, nil
}

func setupSyncMock(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	syncOnce.Do(func() {
		var err error
		matcher := sqlmock.QueryMatcherFunc(func(expected, actual string) error {
			syncMu.Lock()
			syncSQL = append(syncSQL, actual)
			syncMu.Unlock()
			return sqlmock.QueryMatcherRegexp.Match(expected, actual)
		})
		syncDB, syncMock, err = sqlmock.New(sqlmock.QueryMatcherOption(matcher))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		if err = gdb.Register("sqlmock_sync", &syncMockDriver{}); err != nil {
			t.Fatalf("register driver: %v", err)
		}
		a, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
		if !ok {
			t.Fatal("unexpected config adapter")
		}
		link := "sqlmock_sync:u:p@tcp(127.0.0.1:3306)/test"
		a.SetContent(`{"database":{"default":{"type":"sqlmock_sync","link":"` + link + `"},"site_t1":{"type":"sqlmock_sync","link":"` + link + `"}}}`)
	})
	syncMu.Lock()
	syncSQL = nil
	syncMu.Unlock()
	t.Cleanup(func() {
		if err := syncMock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}
	})
	return syncMock
}

func executed(prefix string) []string {
	syncMu.Lock()
	defer syncMu.Unlock()
	var out []string
	for _, s := range syncSQL {
		if strings.HasPrefix(s, prefix) {
			out = append(out, s)
		}
	}
	return out
}

func testParams() *platform.GetGameSourceParams {
	return &platform.GetGameSourceParams{SiteId: 1, SiteCode: "t1", GameId: 17}
}

func testGame() map[string]interface{} {
	return map[string]interface{}{
		"code": "G1", "name": "Game 1", "style": "slot", "img": "https://third/a.png",
		"img1": "https://third/a1.png", "language": "zh", "device": "h5", "currency": "CNY", "vendor": "v",
	}
}

// 已存在记录时，ONE 同步只更新三方字段，不得覆盖 cover_url 及后台管理字段。
func TestSyncGamesToDatabase_UpdateKeepsCoverURL(t *testing.T) {
	mock := setupSyncMock(t)
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM `game_list`").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(1)"}).AddRow(1))
	mock.ExpectExec("^UPDATE `game_list` SET ").WillReturnResult(sqlmock.NewResult(0, 1))

	if err := (&OnePlatform{}).syncGamesToDatabase(context.Background(), testParams(), []map[string]interface{}{testGame()}); err != nil {
		t.Fatalf("sync: %v", err)
	}
	ups := executed("UPDATE")
	if len(ups) != 1 {
		t.Fatalf("expected 1 UPDATE, got %v", ups)
	}
	set := regexp.MustCompile("SET (.*) WHERE").FindStringSubmatch(ups[0])[1]
	for _, forbidden := range []string{"cover_url", "alias_name", "hot", "recommend", "is_new", "status", "sort", "like_num"} {
		if strings.Contains(set, "`"+forbidden+"`") {
			t.Errorf("sync UPDATE must not touch %s: %s", forbidden, ups[0])
		}
	}
	for _, want := range []string{"img", "img1", "name"} {
		if !strings.Contains(set, "`"+want+"`=?") {
			t.Errorf("sync UPDATE should set %s (parameterized): %s", want, ups[0])
		}
	}
}

// 新记录插入时也不写 cover_url（使用表默认值 ”）。
func TestSyncGamesToDatabase_InsertHasNoCoverURL(t *testing.T) {
	mock := setupSyncMock(t)
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM `game_list`").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(1)"}).AddRow(0))
	mock.ExpectExec("^INSERT INTO `game_list`").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := (&OnePlatform{}).syncGamesToDatabase(context.Background(), testParams(), []map[string]interface{}{testGame()}); err != nil {
		t.Fatalf("sync: %v", err)
	}
	ins := executed("INSERT")
	if len(ins) != 1 || strings.Contains(ins[0], "cover_url") {
		t.Errorf("unexpected INSERT: %v", ins)
	}
}
