package main

import "github.com/liuminhaw/mm-plugins/utils"

var propsConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAttributesMiner(client, attributes)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newTaggingMiner(client, tagging)
	},
}
