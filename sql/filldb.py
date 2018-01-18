#!/usr/bin/python
# encoding=utf8

import argparse
import logging
import MySQLdb

# 这个脚本用来填充数据库
# 主要是为了方便测试

# 参数解析

parser = argparse.ArgumentParser()
parser.add_argument("--host", default="localhost", help="MySQL host")
parser.add_argument("--port", default=3306, type=int, help="MySQL port")
parser.add_argument("--db", default="AdClickTool", help="MySQL database")
parser.add_argument("--user", default="root", help="MySQL user")
parser.add_argument("--password", default="123456", help="MySQL password")
parser.add_argument("--loglevel", default="DEBUG", help="ERROR,WARN,INFO,DEBUG")
parser.add_argument("--cleardb", default=False, type=bool, help="clear database")
parser.add_argument("--idtext", default="aabbcc", help="User.idText")
parser.add_argument("--campdomain", default="mymac", help="campaign url domain")

args = parser.parse_args()
logging.basicConfig(level=args.loglevel)


logging.debug("connectin MySQL...")
conn = MySQLdb.Connect(host=args.host, port=args.port, user=args.user, passwd=args.password, db=args.db)
logging.debug("MySQL connected")

print args.cleardb
if args.cleardb:
    clear_tables = [
        ("User", "id"),
        ("AffiliateNetwork", "id"),
        ("Offer", "id"),
        ("Lander", "id"),
        ("Path", "id"),
        ("Lander2Path", "pathId"),
        ("Offer2Path", "pathId"),
        ("Rule", "id"),
        ("Flow", "id"),
        ("Rule2Flow", "flowId"),
        ("TrafficSource", "id"),
        ("TrackingCampaign", "id"),

        ("User", "id"),
        ("User", "id"),
        ("User", "id"),
        ("User", "id"),
        ("User", "id"),
    ]
    cursor = conn.cursor()
    for table, where in clear_tables:
        cursor.execute("DELETE FROM %s WHERE `%s`!=0" % (table, where))
    conn.commit()

# 执行数据库

# user表
cursor = conn.cursor()
idText = args.idtext
print type(idText)
cursor.execute("INSERT INTO User (idText, email, password, firstname, lastname, rootdomainredirect, json, setting,referralToken,contact)"
               " VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (idText, 'tempusage@yeah.net', 'unknown', 'firstname', 'lastname', 'http://www.newbidder.com/', 'json', 'setting','referralToken','contact'))
conn.commit()
userId = cursor.lastrowid
print "userId", userId


################################################################################
# 创建 AffiliateNetwork
################################################################################

postabck_url = "http://%s.%s/postback?cid=REPLACE&payout=OPTIONAL&txid=OPTIONAL"%(idText, args.campdomain)

cursor.execute("INSERT INTO AffiliateNetwork (userId, name, hash, postbackUrl, appendClickId, duplicatedPostback, ipWhiteList)"
               " VALUES (%s,%s,%s,%s,%s,%s,%s)",
               (userId, "AffiliateNetwork.test1", "12345678", postabck_url, 0, 1, """192.168.0.1"""),)
conn.commit()

affiliateId1 = cursor.lastrowid
print "affiliateId1", affiliateId1

cursor.execute("INSERT INTO AffiliateNetwork (userId, name, hash, postbackUrl, appendClickId, duplicatedPostback, ipWhiteList)"
               " VALUES (%s,%s,%s,%s,%s,%s,%s)",
               (userId, "AffiliateNetwork.test2", "12345678", postabck_url, 0, 1, """192.168.0.1"""),)
conn.commit()
affiliateId2 = cursor.lastrowid
print "affiliateId2", affiliateId2

################################################################################
# 创建 Offer
################################################################################

offer_url = "http://download.androidapp.baidu.com/public/uploads/store_0/static/mobisummer4_gl_bd_apk-com.duapps.cleaner-1220_mobisummer4_gl_bd_apk.apk?cid={click.id}"

offerPostbackUrl = "http://zx1jg.voluumtrk2.com/postback?cid=REPLACE&payout=OPTIONAL&txid=OPTIONAL"

cursor.execute("INSERT INTO Offer (userId, name, hash, url, country, AffiliateNetworkId, AffiliateNetworkName, postbackUrl, payoutMode, payoutValue)"
               " VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "Offer 1", "offer.hash", offer_url, "CHN", affiliateId1, "affiliateName1", offerPostbackUrl, 1, 1.0),)
offer1 = cursor.lastrowid

cursor.execute("INSERT INTO Offer (userId, name, hash, url, country, AffiliateNetworkId, AffiliateNetworkName, postbackUrl, payoutMode, payoutValue)"
               " VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "Offer 2", "offer.hash", offer_url, "CHN", affiliateId2, "affiliateName2", offerPostbackUrl, 0, 0.0),)
offer2 = cursor.lastrowid

print "offer1", offer1
print "offer2", offer2

################################################################################
# 创建 Lander
################################################################################

cursor.execute("INSERT INTO Lander (userId, name, hash, url, country, numberOfOffers) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "Lander 1", "lander.hash", "http://lander.com/my/lander/page/1/?cid={click.id}", "CHN", 1),)
lander1 = cursor.lastrowid

cursor.execute("INSERT INTO Lander (userId, name, hash, url, country, numberOfOffers) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "Lander 2", "lander.hash", "http://lander.com/my/lander/page/2/?cid={click.id}", "CHN", 3),)
lander2 = cursor.lastrowid


################################################################################
# 创建 Path
################################################################################

# path1
cursor.execute("INSERT INTO Path (userId, name, hash, redirectMode, directLink, status) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "path.1", "path.1.hash", 0, 0, 1))
path1 = cursor.lastrowid

cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander1, path1, 50))
cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander2, path1, 50))

# path2
cursor.execute("INSERT INTO Path (userId, name, hash, redirectMode, directLink, status) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "path.2", "path.2.hash", 1, 1, 1))
path2 = cursor.lastrowid

cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander1, path2, 100))
cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander2, path2, 400))


# path3
cursor.execute("INSERT INTO Path (userId, name, hash, redirectMode, directLink, status) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "path.3", "path.3.hash", 2, 1, 1))
path3 = cursor.lastrowid

cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander1, path3, 100))
cursor.execute("INSERT INTO Lander2Path (landerId, pathId, weight) VALUES(%s,%s,%s)",
               (lander2, path3, 400))


#
cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer1, path1, 100, 0))
cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer2, path1, 100, 0))

cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer1, path2, 100, 0))
cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer2, path2, 100, 0))

cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer1, path3, 100, 0))
cursor.execute("INSERT INTO Offer2Path (offerId, pathId, weight, deleted) VALUES(%s,%s,%s,%s)", (offer2, path3, 100, 0))


################################################################################
# 创建 Rule
################################################################################

cursor.execute("INSERT INTO Rule (userId, name, hash, type, json, status) VALUES(%s,%s,%s,%s,%s,%s)",
               (userId, "rule.1", "rule.1.hash", 0, "[]", 1))
rule1 = cursor.lastrowid

cursor.execute("INSERT INTO Rule (userId, name, hash, type, json, status) VALUES(%s,%s,%s,%s,%s,%s)",
               (userId, "rule.2", "rule.2.hash", 1, "[]", 1))
rule2 = cursor.lastrowid

# 每个rule都有两个path
cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path1, rule1, 100, 1))

cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path2, rule1, 100, 1))

cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path3, rule1, 100, 1))

cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path1, rule2, 100, 1))

cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path2, rule2, 100, 1))

cursor.execute("INSERT INTO Path2Rule (pathId, ruleId, weight, status) VALUES(%s,%s,%s,%s)",
    (path3, rule2, 100, 1))

################################################################################
# 创建 Flow
################################################################################

cursor.execute("INSERT INTO Flow (userId, name, hash, country, type, redirectMode) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "flow.1", "flow.1.hash", "CHN", 0, 0))
flow1 = cursor.lastrowid

cursor.execute("INSERT INTO Flow (userId, name, hash, country, type, redirectMode) VALUES (%s,%s,%s,%s,%s,%s)",
               (userId, "flow.2", "flow.2.hash", "CHN", 0, 0))
flow2 = cursor.lastrowid

# 每一个flow都有两个rule
cursor.execute("INSERT INTO Rule2Flow (ruleId, flowId, status) VALUES (%s,%s,%s)",
               (rule1, flow1, 1))
cursor.execute("INSERT INTO Rule2Flow (ruleId, flowId, status) VALUES (%s,%s,%s)",
               (rule2, flow1, 1))

cursor.execute("INSERT INTO Rule2Flow (ruleId, flowId, status) VALUES (%s,%s,%s)",
               (rule1, flow2, 1))
cursor.execute("INSERT INTO Rule2Flow (ruleId, flowId, status) VALUES (%s,%s,%s)",
               (rule2, flow2, 1))

################################################################################
# 创建 TrafficSource
################################################################################


traffic_postback = "http://freeapi.ipip.net/8.8.8.8"
params="""[
  {
    "Parameter": "ADBLOCK",
    "PlaceHolder": "[ADBLOCK]",
    "Name": "ADBLOCK",
    "Track": 0
  },
  {
    "Parameter": "ISPNAME",
    "PlaceHolder": "[ISPNAME]",
    "Name": "ISPNAME",
    "Track": 0
  },
  {
    "Parameter": "SCREENRESOLUTION",
    "PlaceHolder": "[SCREENRESOLUTION]",
    "Name": "SCREENRESOLUTION",
    "Track": 0
  }
]"""

cursor.execute("INSERT INTO TrafficSource (userId, name, hash, postbackUrl, pixelRedirectUrl, impTracking, externalId, campaignId, websiteId,"
               " cost, params) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "TrafficSource.1", "TrafficSource.1.hash", traffic_postback,
                "http://traffic.source.com/pixel/redirect/url/1/", 0, "{}", "{}", "{}", "{}", params))
traffic1 = cursor.lastrowid

cursor.execute("INSERT INTO TrafficSource (userId, name, hash, postbackUrl, pixelRedirectUrl, impTracking, externalId, campaignId, websiteId,"
               " cost, params) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "TrafficSource.2", "TrafficSource.2.hash", traffic_postback,
                "http://traffic.source.com/pixel/redirect/url/2/", 0, "{}", "{}", "{}", "{}", params))
traffic2 = cursor.lastrowid


################################################################################
# 创建 Campaign
################################################################################

cursor.execute("INSERT INTO TrackingCampaign (userId, name, hash, url, impPixelUrl, trafficSourceId, trafficSourceName,"
               " country, costModel, cpcValue, cpaValue, cpmValue, redirectMode, targetType, targetFlowId, targetUrl, "
               "status) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "campaign.1", "463d-be39-e1f51d80591", "http://%s.%s/campaign.1/"%(idText, args.campdomain),
                "http://%s.%s/impression/campaign.1/"%(idText, args.campdomain), traffic1, "TrafficSource.1",
                "CHN", 1, 1.1, 1.2, 1.3, 0, 1, flow1, "", 1,))
campaign1 = cursor.lastrowid


cursor.execute("INSERT INTO TrackingCampaign (userId, name, hash, url, impPixelUrl, trafficSourceId, trafficSourceName,"
               " country, costModel, cpcValue, cpaValue, cpmValue, redirectMode, targetType, targetFlowId, targetUrl, "
               "status) VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)",
               (userId, "campaign.2", "463d-be39-e1f51d80592", "http://%s.%s/campaign.2/"%(idText, args.campdomain),
                "http://%s.%s/impression/campaign.2/"%(idText, args.campdomain), traffic1, "TrafficSource.1",
                "CHN", 2, 1.1, 1.2, 1.3, 0, 1, flow2, "", 1,))
campaign2 = cursor.lastrowid

conn.commit()
