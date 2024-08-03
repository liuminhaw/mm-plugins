package main

import (
	"github.com/liuminhaw/mm-plugins/utils"
)

var propsConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRegionMiner(location)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAccelerateMiner(client, accelerateConfig)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAnalyticsMiner(client, analyticsConfig)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAclMiner(client, acl)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newCorsMiner(client, cors)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newEncryptionMiner(client, encryption)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newIntelligentTieringMiner(client, intelligentTiering)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newInventoryMiner(client, inventory)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newLifecycleMiner(client, lifecycle)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newLoggingMiner(client, logging)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newMetricsMiner(client, metrics)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newNotificationMiner(client, notification)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newOwnershipControlMiner(client, ownershipControl)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newPolicyMiner(client, policy)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newPolicyStatusMiner(client, policyStatus)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newReplicationMiner(client, replication)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRequestPaymentMiner(client, requestPayment)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newTaggingMiner(client, tagging)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newVersioningMiner(client, versioning)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newWebsiteMiner(client, website)
	},
}
