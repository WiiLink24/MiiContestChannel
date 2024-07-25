package main

import (
	"context"
	"fmt"
	"github.com/WiiLink24/MiiContestChannel/common"
	"github.com/WiiLink24/MiiContestChannel/contest"
	"github.com/WiiLink24/MiiContestChannel/first"
	"github.com/WiiLink24/MiiContestChannel/plaza"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ctx  = context.Background()
	pool *pgxpool.Pool
)

func main() {
	config := common.GetConfig()

	// Start SQL
	dbString := fmt.Sprintf("postgres://%s:%s@%s/%s", config.Username, config.Password, config.DatabaseAddress, config.DatabaseName)
	dbConf, err := pgxpool.ParseConfig(dbString)
	common.CheckError(err)
	pool, err = pgxpool.ConnectConfig(ctx, dbConf)
	common.CheckError(err)

	err = first.MakeFirst()
	common.CheckError(err)

	err = first.MakeAddition()
	common.CheckError(err)

	err = contest.MakeContestInfos(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeSpotList(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeBargainList(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeNewList(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeTop50(pool, ctx)
	common.CheckError(err)

	err = plaza.MakePopCraftsList(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeNumberInfo(pool, ctx)
	common.CheckError(err)

	err = plaza.MakeSelectList(pool, ctx)
	common.CheckError(err)
}
