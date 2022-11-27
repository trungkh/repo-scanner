package config

import (
	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/sqlq"
)

func (c *Config) InitQuery() serror.SError {
	opt := sqlq.BuilderOption{
		Driver: sqlq.DriverPostgreSQL,
	}

	/*opMaps := sqlq.OperatorsMap{
		"basic": []sqlq.Operator{
			sqlq.OperatorEqual,
			sqlq.OperatorNotEqual,
		},
		"key": []sqlq.Operator{
			sqlq.OperatorEqual,
			sqlq.OperatorNotEqual,
			sqlq.OperatorIn,
			sqlq.OperatorNotIn,
		},
		"number": []sqlq.Operator{
			sqlq.OperatorEqual,
			sqlq.OperatorNotEqual,
			sqlq.OperatorIn,
			sqlq.OperatorNotIn,
			sqlq.OperatorGreater,
			sqlq.OperatorGreaterThen,
			sqlq.OperatorLess,
			sqlq.OperatorLessThen,
			sqlq.OperatorBetween,
		},
		"text": []sqlq.Operator{
			sqlq.OperatorEqual,
			sqlq.OperatorNotEqual,
			sqlq.OperatorLike,
			sqlq.OperatorNotLike,
			sqlq.OperatorILike,
			sqlq.OperatorNotILike,
		},
		"nullable": []sqlq.Operator{
			sqlq.OperatorEqual,
			sqlq.OperatorNotEqual,
			sqlq.OperatorIsNull,
			sqlq.OperatorIsNotNull,
		},
	}*/

	/*most(opt.Tables.AddFromStruct("influencers", sqlq.NewTableOption{
		Schema:       "igscraper",
		Table:        "influencers",
		ConditionMap: opMaps,
	}, model.Influencers{}))

	most(opt.Tables.AddFromStruct("images", sqlq.NewTableOption{
		Schema:       "igscraper",
		Table:        "influencer_images",
		ConditionMap: opMaps,
	}, model.InfluencerImages{}))

	most(opt.Tables.AddFromStruct("status", sqlq.NewTableOption{
		Schema:       "igscraper",
		Table:        "scrape_status",
		ConditionMap: opMaps,
	}, model.ScrapeStatus{}))*/

	c.Query = sqlq.NewBuilder(opt)
	return nil
}
