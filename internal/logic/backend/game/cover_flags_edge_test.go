package game

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	v1 "jh_game_service/api/backend/game/v1"
)

// 显式 0 写入，-1 不写入；混合场景只更新显式值
func TestUpdateGameListItem_MixedZeroAndMinusOne(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	mock.ExpectExec("^UPDATE `game_list` SET .* WHERE id = \\? AND site_id = \\?$").
		WillReturnResult(sqlmock.NewResult(0, 1))

	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: 0, Recommend: -1, IsNew: 0, Status: -1, Sort: -1,
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	f := setFields(lastUpdateSQL(t))
	if !f["hot"] || !f["is_new"] || f["recommend"] || f["status"] || f["sort"] || f["cover_url"] {
		t.Errorf("unexpected SET fields: %v", f)
	}
}

// sort=0 有效；全部 -1 且无封面时不发 UPDATE
func TestUpdateGameListItem_SortZeroVsMinusOne(t *testing.T) {
	mock := setupCoverMock(t)
	expectExists(mock, 1)
	mock.ExpectExec("^UPDATE `game_list` SET .* WHERE id = \\? AND site_id = \\?$").
		WillReturnResult(sqlmock.NewResult(0, 1))
	res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: 0,
	})
	if err != nil || !res.Success {
		t.Fatalf("expected success, res=%v err=%v", res, err)
	}
	if f := setFields(lastUpdateSQL(t)); !f["sort"] {
		t.Errorf("expected sort in SET, got %v", f)
	}

	mock2 := setupCoverMock(t)
	expectExists(mock2, 1)
	res, err = (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
		Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: -1,
	})
	if err != nil || res.Success {
		t.Fatalf("expected 'no fields' failure, res=%v err=%v", res, err)
	}
}

// 非法 cover_url 在 UPDATE 前失败，不写库
func TestUpdateGameListItem_InvalidCoverRejected(t *testing.T) {
	for _, u := range []string{
		"javascript:alert(1)",
		"data:image/png;base64,AAAA",
		"file:///etc/passwd",
		"https://",
		"https://a.com/" + strings.Repeat("x", 500),
		"https://cdn.example.com@evil.com/a.png",
	} {
		mock := setupCoverMock(t)
		expectExists(mock, 1)
		res, err := (&sGame{}).UpdateGameListItem(context.Background(), &v1.UpdateGameListItemReq{
			Id: 5, SiteId: 1, Hot: -1, Recommend: -1, IsNew: -1, Status: -1, Sort: -1, CoverUrl: u,
		})
		if err != nil || res.Success {
			t.Errorf("%q: expected failure, res=%v err=%v", u, res, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("%q: unexpected DB call: %v", u, err)
		}
	}
}
