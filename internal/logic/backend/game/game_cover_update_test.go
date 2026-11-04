package game

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	v1 "jh_game_service/api/backend/game/v1"
)

// coverMockDriver 复用 GoFrame MySQL 方言，底层连接替换为 sqlmock。
type coverMockDriver struct{ *mysql.Driver }

var (
	coverOnce sync.Once
	coverDB   *sql.DB
	coverMock sqlmock.Sqlmock
	coverMu   sync.Mutex
	coverSQL  []string // 记录实际执行的 SQL 文本
)

func (d *coverMockDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	return &coverMockDriver{Driver: &mysql.Driver{Core: core}}, nil
}

func (d *coverMockDriver) Open(config *gdb.ConfigNode) (*sql.DB, error) { return coverDB, nil }

// TableFields 返回 game_list 固定表结构，避免 SHOW FULL COLUMNS 查询。
func (d *coverMockDriver) TableFields(ctx context.Context, table string, schema ...string) (map[string]*gdb.TableField, error) {
	names := []string{"id", "site_id", "type", "platform", "name", "alias_name", "style", "img", "img1", "device", "vendor", "code", "language", "hot", "recommend", "like_num", "is_new", "sort", "status", "available", "created_at", "updated_at", "game_id", "cover_url"}
	m := make(map[string]*gdb.TableField, len(names))
	for i, n := range names {
		m[n] = &gdb.TableField{Index: i, Name: n}
	}
	return m, nil
}

// recordingMatcher 正则匹配并记录实际 SQL。
var recordingMatcher = sqlmock.QueryMatcherFunc(func(expected, actual string) error {
	coverMu.Lock()
	coverSQL = append(coverSQL, actual)
	coverMu.Unlock()
	return sqlmock.QueryMatcherRegexp.Match(expected, actual)
})

func setupCoverMock(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	coverOnce.Do(func() {
		var err error
		coverDB, coverMock, err = sqlmock.New(sqlmock.QueryMatcherOption(recordingMatcher))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		if err = gdb.Register("sqlmock_cover", &coverMockDriver{}); err != nil {
			t.Fatalf("register driver: %v", err)
		}
		if a, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			a.SetContent(`{"database":{"default":{"type":"sqlmock_cover","link":"sqlmock_cover:u:p@tcp(127.0.0.1:3306)/test"}}}`)
		} else {
			t.Fatal("unexpected config adapter")
		}
	})
	coverMu.Lock()
	coverSQL = nil
	coverMu.Unlock()
	t.Cleanup(func() {
		if err := coverMock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}
	})
	return coverMock
}

// recArgs 记录绑定参数。
type recArgs struct{ got *[]driver.Value }

func (r recArgs) Match(v driver.Value) bool { *r.got = append(*r.got, v); return true }

func nArgs(n int, rec *[]driver.Value) []driver.Value {
	out := make([]driver.Value, n)
	for i := range out {
		out[i] = recArgs{rec}
	}
	return out
}

func hasArg(args []driver.Value, v string) bool {
	for _, a := range args {
		if s, ok := a.(string); ok && s == v {
			return true
		}
	}
	return false
}

func expectExists(mock sqlmock.Sqlmock, n int) {
	mock.ExpectQuery("SELECT COUNT\\(1\\) FROM `game_list`").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(1)"}).AddRow(n))
}

func lastUpdateSQL(t *testing.T) string {
	t.Helper()
	coverMu.Lock()
	defer coverMu.Unlock()
	for i := len(coverSQL) - 1; i >= 0; i-- {
		if strings.HasPrefix(coverSQL[i], "UPDATE") {
			return coverSQL[i]
		}
	}
	t.Fatal("no UPDATE executed")
	return ""
}

// setFields 解析 UPDATE ... SET a=?,b=? WHERE 中的字段名。
func setFields(sqlText string) map[string]bool {
	m := regexp.MustCompile("SET (.*) WHERE").FindStringSubmatch(sqlText)
	out := map[string]bool{}
	if m == nil {
		return out
	}
	for _, f := range regexp.MustCompile("`(\\w+)`=\\?").FindAllStringSubmatch(m[1], -1) {
		out[f[1]] = true
	}
	return out
}

func TestUpdateGameListItem_SetCoverURL(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	mock.ExpectExec("^UPDATE `game_list` SET .* WHERE id = \\? AND site_id = \\?$").
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: 1, Sort: -1,
		CoverUrl: "  https://cdn.example.com/uploads/a.png  ",
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	f := setFields(lastUpdateSQL(t))
	if !f["cover_url"] || !f["status"] || f["hot"] || f["img"] || f["img1"] {
		t.Errorf("unexpected SET fields: %v", f)
	}
}

func TestUpdateGameListItem_CoverOnlyWithMinusOneFlags(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	var args []driver.Value
	mock.ExpectExec("^UPDATE `game_list` SET (`cover_url`=\\?,`updated_at`=\\?|`updated_at`=\\?,`cover_url`=\\?) WHERE id = \\? AND site_id = \\?$").
		WithArgs(nArgs(4, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: -1,
		CoverUrl: "  https://cdn.example.com/uploads/a.png  ",
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	if !hasArg(args, "https://cdn.example.com/uploads/a.png") {
		t.Errorf("cover_url should be trimmed and bound as parameter, args=%#v", args)
	}
}

func TestUpdateGameListItem_ClearCover(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	var args []driver.Value
	mock.ExpectExec("^UPDATE `game_list` SET (`cover_url`=\\?,`updated_at`=\\?|`updated_at`=\\?,`cover_url`=\\?) WHERE id = \\? AND site_id = \\?$").
		WithArgs(nArgs(4, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))

	// clear_cover 优先于 cover_url
	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: -1,
		CoverUrl: "https://cdn.example.com/x.png", ClearCover: 1,
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	if !hasArg(args, "") || hasArg(args, "https://cdn.example.com/x.png") {
		t.Errorf("clear_cover should write empty cover_url, args=%#v", args)
	}
}

// 服务端把 >=0 的值写入；proto 默认值 0 会写成 0。网关已把“未传”转成 -1，本用例模拟未走网关、直接传零值。
func TestUpdateGameListItem_OmittedFlagsWrittenAsZero(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	mock.ExpectExec("^UPDATE `game_list` SET .* WHERE id = \\? AND site_id = \\?$").
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, CoverUrl: "https://cdn.example.com/a.png",
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	f := setFields(lastUpdateSQL(t))
	for _, k := range []string{"hot", "recommend", "is_new", "status", "sort", "cover_url"} {
		if !f[k] {
			t.Errorf("expected %s in SET, got %v", k, f)
		}
	}
	if f["img"] || f["img1"] || f["alias_name"] {
		t.Errorf("unexpected fields in SET: %v", f)
	}
}

func TestUpdateGameListItem_BlankCoverIgnored(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: -1, CoverUrl: "   ",
	})
	if err != nil || res.Success {
		t.Fatalf("expected 'no fields' failure without UPDATE, res=%v err=%v", res, err)
	}
	_ = mock // 不应发出 UPDATE
}

func TestUpdateGameListItem_NotFound(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 0)
	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 999, SiteId: 1, CoverUrl: "https://cdn.example.com/a.png",
	})
	if err != nil || res.Success {
		t.Fatalf("expected not-found failure, res=%v err=%v", res, err)
	}
}

func TestEffectiveImg(t *testing.T) {
	cases := []struct{ cover, third, want string }{
		{"https://c/a.png", "https://t/b.png", "https://c/a.png"},
		{"", "https://t/b.png", "https://t/b.png"},
		{"   ", "https://t/b.png", "https://t/b.png"},
		{"", "", ""},
	}
	for _, c := range cases {
		if got := effectiveImg(c.cover, c.third); got != c.want {
			t.Errorf("effectiveImg(%q,%q)=%q want %q", c.cover, c.third, got, c.want)
		}
	}
}
