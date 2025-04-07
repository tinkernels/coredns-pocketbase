package pb_migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[1-9][0-9]{17}",
					"hidden": false,
					"id": "text3208210256",
					"max": 18,
					"min": 1,
					"name": "id",
					"pattern": "^[1-9][0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text2699804679",
					"max": 0,
					"min": 0,
					"name": "zone",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text1579384326",
					"max": 0,
					"min": 0,
					"name": "name",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "number2750318623",
					"max": null,
					"min": null,
					"name": "ttl",
					"onlyInt": true,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "json4274335913",
					"maxSize": 0,
					"name": "content",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text1768539901",
					"max": 0,
					"min": 0,
					"name": "record_type",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_186858105",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_nRNNJl6OY9` + "`" + ` ON ` + "`" + `coredns_records` + "`" + ` (\n  ` + "`" + `zone` + "`" + `,\n  ` + "`" + `name` + "`" + `,\n  ` + "`" + `record_type` + "`" + `\n)"
			],
			"listRule": null,
			"name": "coredns_records",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_186858105")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
