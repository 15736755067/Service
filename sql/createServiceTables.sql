CREATE TABLE AdClickTool.`User` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `idText` varchar(8) NOT NULL DEFAULT '' COMMENT '用在click,postback等url中，用于区别用户',
  `email` varchar(50) NOT NULL,
  `emailVerified` int(11) NOT NULL DEFAULT 0 COMMENT '0:邮箱未验证;1:邮箱已验证',
  `contact` text NOT NULL DEFAULT '' COMMENT '用户联系方式的json格式内容,{"type":"skype","value":"lijintao"}',
  `password` varchar(256) NOT NULL DEFAULT '',
  `firstname` varchar(256) NOT NULL DEFAULT '',
  `lastname` varchar(256) NOT NULL DEFAULT '',
  `campanyName` varchar(256) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0:New;1:运行中;2:已过期;3:Events已消耗完（包括透支）;4:定制套餐已过期',
  `registerts` int(11) COMMENT '用户的注册时间的时间戳',
  `lastLogon` int(11) COMMENT '上次登录时间的时间戳',
  `timezone` varchar(6) DEFAULT '+00:00' COMMENT '默认的时区，格式:+08:00',
  `timezoneId` int(11) DEFAULT '35' COMMENT '用户的时区Id，默认为GMT',
  `rootdomainredirect` varchar(512) NOT NULL DEFAULT '' COMMENT '当访问用户的rootdomain时的跳转页面，如果为空则显示默认的404页面',
  `json` text NOT NULL COMMENT '按照既定规则生成的User信息(CompanyName,Phone,DefaultTimeZone,DefaultHomeScreen)',
  `setting` text NOT NULL DEFAULT '',
  `referralToken` varchar(128) NOT NULL COMMENT '用户推荐链接中的token，链接中其他部分现拼',
  `currentGroup` varchar(128) DEFAULT '' COMMENT '用户当前选择的Group的Id，为空则表示为自己为Owner的Group',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idText` (`idtext`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;


CREATE TABLE AdClickTool.`UserGroup` (
	`id` int(11) NOT NULL AUTO_INCREMENT,
	`groupId` varchar(128) NOT NULL,
	`userId` int(11) NOT NULL,
	`role` int(11) NOT NULL DEFAULT 0 COMMENT '用户所属权限，0:owner;1:member',
	`privilege` text NOT NULL COMMENT '角色权限的config信息(json)',
	`createdAt` int(11) NOT NULL COMMENT '用户加入Group的时间戳',
	`deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
	PRIMARY KEY (`id`),
	UNIQUE KEY `unique_key` (`groupId`,`userId`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`GroupInvitation` (
	`id` int(11) NOT NULL AUTO_INCREMENT,
	`userId` int(11) NOT NULL COMMENT '发出邀请的userId',
	`groupId` varchar(128) NOT NULL COMMENT '邀请进入的groupId',
	`inviteeEmail` varchar(50) NOT NULL COMMENT '被邀请者的Email',
	`inviteeTime` int(11) NOT NULL COMMENT '邀请发生的时间',
	`status` int(11) NOT NULL DEFAULT 0 COMMENT '状态，0:新建;1:接受邀请;2:拒绝邀请;3:取消邀请',
	`code` varchar(39) NOT NULL COMMENT '邀请码',
	`deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
	PRIMARY KEY (`id`),
	UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserDomain` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `domain` varchar(512) NOT NULL DEFAULT '' COMMENT '用于生成campaignurl和clickurl时用到的域名',
  `main` int(11) NOT NULL DEFAULT 0 COMMENT '0:非主域名;1:主域名',
  `verified` int(11) NOT NULL DEFAULT 0 COMMENT '域名是否验证过，0:未验证;1:已验证',
  `customize` int(11) NOT NULL DEFAULT 0 COMMENT '0:系统默认的域名;1:用户添加的域名',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`TrackingCampaign` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `url` varchar(512) NOT NULL COMMENT '根据campaign内容生成的tracking url(不包含参数部分)',
  `impPixelUrl` varchar(512) NOT NULL COMMENT '根据campaign内容生成的tracking url(不包含参数部分)',
  `trafficSourceId` int(11) NOT NULL,
  `trafficSourceName` varchar(256) NOT NULL DEFAULT '',
  `country` varchar(3) NOT NULL DEFAULT '' COMMENT 'ISO-ALPHA-3',
  `costModel` int(11) NOT NULL COMMENT '0:Do-not-track-costs;1:cpc;2:cpa;3:cpm;4:auto?',
  `cpcValue` decimal(10,5) NOT NULL DEFAULT 0,
  `cpaValue` decimal(10,5) NOT NULL DEFAULT 0,
  `cpmValue` decimal(10,5) NOT NULL DEFAULT 0,
  `postbackUrl` varchar(512) NOT NULL DEFAULT '' COMMENT 'campaign自定义的postback url，为空则使用traffic source的设置',
  `pixelRedirectUrl` varchar(512) NOT NULL DEFAULT '' COMMENT 'campaign自定义的pixel redirect url，为空则使用traffic source的设置',
  `redirectMode` int(11) NOT NULL COMMENT '0:302;1:Meta refresh;2:Double meta refresh',
  `targetType` int(11) NOT NULL DEFAULT 0 COMMENT '跳转类型,0:URL;1:Flow;2:Rule;3:Path;4:Lander;5:Offer',
  `targetFlowId` int(11) NOT NULL,
  `targetUrl` varchar(512) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL COMMENT '0:停止;1:运行',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Campaign的配置信息';

CREATE TABLE AdClickTool.`Flow` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `country` varchar(3) NOT NULL DEFAULT '' COMMENT 'ISO-ALPHA-3',
  `type` int(11) NOT NULL COMMENT '0:匿名;1:普通(标示Campaign里选择Flow时是否可见)',
  `redirectMode` int(11) NOT NULL COMMENT '0:302;1:Meta refresh;2:Double meta refresh',
  `createdAt` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Flow的配置信息';

CREATE TABLE AdClickTool.`Rule` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `type` int(11) NOT NULL COMMENT '0:匿名;1:普通(标示是否是Flow里默认Path的Rule)',
  `json` text NOT NULL COMMENT '按照既定规则生成的rule信息，供Service使用',
  `object` text NOT NULL COMMENT '按照既定规则生成的rule信息，供前端使用',
  `status` int(11) NOT NULL COMMENT '0:停止;1:运行;用来标记该Rule本身是否有效',
  `createdAt` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Rule的配置信息';

CREATE TABLE AdClickTool.`Rule2Flow` (
  `ruleId` int(11) NOT NULL COMMENT '必须非0',
  `flowId` int(11) NOT NULL COMMENT '必须非0',
  `status` int(11) NOT NULL COMMENT '0:停止;1:运行;用来标记Rule在特定Flow中是否有效',
  `order` int(11) NOT NULL DEFAULT 0 COMMENT 'rule在同一个flow内的排列顺序,0~,越小越靠前',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`flowId`,`ruleId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Path` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) NOT NULL,
  `hash` varchar(39) NOT NULL,
  `redirectMode` int(11) NOT NULL COMMENT '0:302;1:Meta refresh;2:Double meta refresh',
  `directLink` int(11) NOT NULL COMMENT '0:No;1:Yes',
  `status` int(11) NOT NULL COMMENT '0:停止;1:运行;用来标记该Path本身是否有效',
  `createdAt` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Path的配置信息';

CREATE TABLE AdClickTool.`Path2Rule` (
  `pathId` int(11) NOT NULL COMMENT '必须非0',
  `ruleId` int(11) NOT NULL COMMENT '必须非0',
  `weight` int(11) NOT NULL COMMENT '0~100',
  `status` int(11) NOT NULL COMMENT '0:停止;1:运行;用来标记Path在特定Rule中是否有效',
  `order` int(11) NOT NULL DEFAULT 0 COMMENT 'path在同一个rule内的排列顺序,0~,越小越靠前',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`ruleId`,`pathId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Lander` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `url` varchar(512) NOT NULL,
  `country` varchar(3) NOT NULL DEFAULT '' COMMENT 'ISO-ALPHA-3',
  `numberOfOffers` int(11) NOT NULL,
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Lander的配置信息';

CREATE TABLE AdClickTool.`Lander2Path` (
  `landerId` int(11) NOT NULL COMMENT '必须非0',
  `pathId` int(11) NOT NULL COMMENT '必须非0',
  `weight` int(11) NOT NULL COMMENT '0~100',
  `order` int(11) NOT NULL DEFAULT 0 COMMENT 'lander在同一个path内的排列顺序,0~,越小越靠前',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`landerId`,`pathId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Offer` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `url` varchar(512) NOT NULL,
  `country` varchar(1024) NOT NULL DEFAULT '' COMMENT 'ISO-ALPHA-3,USA,CHN,IDA,BRA...',
  `thirdPartyOfferId` text DEFAULT '' COMMENT '第三方AffiliateNetwork导入的Offer的Id',
  `AffiliateNetworkId` int(11) NOT NULL COMMENT '标记属于哪家AffiliateNetwork',
  `AffiliateNetworkName` varchar(256) NOT NULL,
  `postbackUrl` varchar(512) NOT NULL,
  `payoutMode` int(11) NOT NULL COMMENT '0:Auto;1:Manual',
  `payoutValue` decimal(10,5) NOT NULL DEFAULT 0,
  `duplicatedClick` int(11) NOT NULL DEFAULT 0 COMMENT '是否允许记录重复的Click，0:不允许;1:允许',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`) 
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户生成的每个Lander的配置信息';

CREATE TABLE AdClickTool.`Offer2Path` (
  `offerId` int(11) NOT NULL COMMENT '必须非0',
  `pathId` int(11) NOT NULL COMMENT '必须非0',
  `weight` int(11) NOT NULL COMMENT '0~100',
  `order` int(11) NOT NULL DEFAULT 0 COMMENT 'offer在同一个path内的排列顺序,0~,越小越靠前',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`offerId`,`pathId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Tags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) NOT NULL,
  `type` int(11) NOT NULL COMMENT '1:Campaign;2:Lander;3:Offer',
  `targetId` int(11) NOT NULL COMMENT '必须非0，根据type不同可能为Campaign/Lander/Offer的Id',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`TrafficSource` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `postbackUrl` varchar(512) NOT NULL,
  `pixelRedirectUrl` varchar(512) NOT NULL,
  `impTracking` int(11) NOT NULL COMMENT '0:No;1:Yes',
  `externalId` varchar(256) NOT NULL COMMENT '按照既定规则生成的ExternalId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `cost` varchar(124) NOT NULL COMMENT '按照既定规则生成的Cost params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `campaignId` varchar(256) NOT NULL COMMENT '按照既定规则生成的CampaignId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `websiteId` varchar(256) NOT NULL COMMENT '按照既定规则生成的WebsiteId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `params` text NOT NULL COMMENT '按照既定规则生成的params信息:[{"Parameter":"X","Placeholder":"X","Name":"X","Track":N(0,1)},...]',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`TemplateTrafficSource` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `order` int(11) NOT NULL COMMENT '序号，越小越靠前',
  `name` varchar(256) NOT NULL,
  `postbackUrl` varchar(512) NOT NULL DEFAULT '',
  `pixelRedirectUrl` varchar(512) NOT NULL DEFAULT '',
  `externalId` varchar(256) NOT NULL COMMENT '按照既定规则生成的ExternalId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `cost` varchar(124) NOT NULL COMMENT '按照既定规则生成的Cost params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `campaignId` varchar(256) NOT NULL COMMENT '按照既定规则生成的CampaignId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `websiteId` varchar(256) NOT NULL COMMENT '按照既定规则生成的WebsiteId params信息:{"Parameter":"X","Placeholder":"X","Name":"X"}',
  `params` text NOT NULL COMMENT '按照既定规则生成的params信息:[{"Parameter":"X","Placeholder":"X","Name":"X","Track":N(0,1)},...]',
  `apiName` varchar(256) NOT NULL COMMENT 'api拉取时，区分用',
  `apiReport` int(11) NOT NULL DEFAULT 0 COMMENT '0:不支持api拉取Report;1:支持拉取Report',
  `apiUrl` text NOT NULL COMMENT 'api拉取Report用的Url',
  `apiParams` text NOT NULL COMMENT 'api参数字段名的json',
-- TODO Task支持的最大时间间隔；Task支持的最早报表开始时间
  `apiMaxTimeSpan` int(11) NOT NULL DEFAULT 0 COMMENT '获取报表的最大时间区间，单位：秒，0表示无限制',
  `apiEarliestTime` int(11) NOT NULL DEFAULT 0 COMMENT '获取报表的最早起始时间点到当前时间的间隔，单位：秒，0表示无限制',
  `apiMode` int(11) NOT NULL DEFAULT 0 COMMENT '1:仅token;2:仅Username/password;3:token/up都支持',
  `apiInterval` int(11) NOT NULL DEFAULT 0 COMMENT '连续两次Task之间的最小间隔时间，0表示没有限制，单位：秒',
  `apiDimensions` text NOT NULL DEFAULT '' COMMENT 'api获取报表时可支持的维度，格式：{"campaignId":"CampaignId","webSiteId":"WebSiteId","v1":"Country","v3":"OS"}',
  `apiMeshSize` varchar(64) NOT NULL DEFAULT '' COMMENT 'api获取报告支持的所有粒度，minute,hour,day,week,month,year',
  `apiTimezones` text NOT NULL DEFAULT '' COMMENT 'api获取报表的可支持的timezone列表，格式：[{id:34,name:"Morocco Standard Time",shift:"+00:00",param:"UTC"},{id:35,name:"UTC",shift:"+00:00",param:"UTC"},{id:36,name:"GMT Standard Time",shift:"+00:00",param:"UTC"}]',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`AffiliateNetwork` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `templateANId` int(11),
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `hash` varchar(39) NOT NULL,
  `postbackUrl` varchar(512) NOT NULL DEFAULT '',
  `appendClickId` int(11) NOT NULL COMMENT '0:No;1:Yes',
  `duplicatedPostback` int(11) DEFAULT 0 NOT NULL COMMENT '0:No;1:Yes',
  `ipWhiteList` text NOT NULL COMMENT 'IP白名单，数组存储成JSON',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`TemplateAffiliateNetwork` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `order` int(11) NOT NULL COMMENT '序号，越小越靠前',
  `name` varchar(256) NOT NULL,
  `postbackParams` text NOT NULL COMMENT '回调url中参数的写法:{cid:%subid1%;p:%commission%}',
  `desc` text NOT NULL COMMENT '关于该AfflicateNetwork的描述，HTML',
  `apiName` varchar(256) NOT NULL COMMENT 'api拉取时，区分用',
  `apiOffer` int(11) NOT NULL COMMENT '0:不支持api拉取Offer;1:支持拉取Offer',
  `apiUrl` text NOT NULL COMMENT 'api拉取Offer用的Url',
  `apiParams` text NOT NULL COMMENT 'api参数字段名的json',
  `apiMode` int(11) NOT NULL DEFAULT 0 COMMENT '1:仅token;2:仅Username/password;3:token/up都支持',
  `apiInterval` int(11) NOT NULL DEFAULT 0 COMMENT '连续两次Task之间的最小间隔时间，0表示没有限制，单位：秒',
  `apiOfferAutoSuffix` varchar(256) NOT NULL DEFAULT '' COMMENT '自动添加在拉取的OfferUrl后面的内容，比如:dv1={click.id}',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UrlTokens` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `token` varchar(64) NOT NULL,
  `desc` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `token` (`token`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Country` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) NOT NULL,
  `alpha2Code` varchar(2) NOT NULL,
  `alpha3Code` varchar(3) NOT NULL,
  `numCode` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `alpha2Code` (`alpha2Code`),
  UNIQUE KEY `alpha3Code` (`alpha3Code`),
  UNIQUE KEY `numCode` (`numCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Timezones` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) NOT NULL,
  `detail` varchar(256) NOT NULL,
  `region` varchar(256) NOT NULL,
  `utcShift` varchar(6) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserReferralLog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `referredUserId` int(11) NOT NULL,
  `acquired` int(11) NOT NULL COMMENT '用户注册的时间，时间戳',
  `status` int(11) NOT NULL COMMENT '0:New刚注册;1:Activated已充值',
  `percent` int(11) NOT NULL COMMENT '提成比例，万分之几，以后挪到单独的表中',
  `lastActivity` int(11) NOT NULL COMMENT '待删除，最近一次充值时间戳',
  `recentCommission` int(11) NOT NULL COMMENT '待删除',
  `totalCommission` int(11) NOT NULL COMMENT '待删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserPlan` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(16) NOT NULL DEFAULT '',
  `userId` int(11) NOT NULL DEFAULT 0 COMMENT '0:通用Plan;>0:特定用户的定制Plan',
  `includedEvents` int(11) NOT NULL DEFAULT -1 COMMENT 'Plan包含的events数，<0:无限制;0:没有events;>0:上限events数',
  `retentionLimit` int(11) NOT NULL DEFAULT 1 COMMENT '报表最长时间月份数',
  `domainLimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持Custom Domain的个数:0~',
  `userLimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持的额外user的个数（不包括自己）:0~',
  `tsReportLimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持的traffic source report的个数:0~',
  `anOfferAPILimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持的affiliate network账号的个数:0~',
  `ffRuleLimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持的fraud filter rule的个数:0~',
  `scRuleLimit` int(11) NOT NULL DEFAULT 0 COMMENT '可支持的sudden change rule的个数:0~',
  `separateIP` int(11) NOT NULL DEFAULT 0 COMMENT '是否有独立IP，0:没有;1:有',
  `price` int(11) NOT NULL DEFAULT 0 COMMENT '价格，单位:USD',
  `hasCommission` int(11) NOT NULL DEFAULT 0 COMMENT '购买该plan是否引入提成，0:没有;1:有',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除', -- 修改plan时，必须通过新建plan来完成
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserBilling` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `planId` int(11) NOT NULL COMMENT '当前的Plan的Id',
  `customPlanId` int(11) NOT NULL DEFAULT 0 COMMENT '用户的当前定制Plan的Id，默认为0，优先使用这个Id',
  `planPaymentLogId` int(11) NOT NULL COMMENT '当前的Plan的购买记录Id',
  `nextPlanId` int(11) NOT NULL DEFAULT 0 COMMENT '下一次的Plan的Id，0则为No Plan',
  `nextPaymentMethod` int(11) NOT NULL DEFAULT 0 COMMENT '下一次的付费方式，0则为No Plan',
  `planStart` int(11) COMMENT 'plan开始时间的时间戳',
  `planEnd` int(11) COMMENT 'plan结束时间的时间戳,<=0:表示没有截止时间',
  `billedEvents` int(11) NOT NULL DEFAULT 0,
  `totalEvents` int(11) NOT NULL DEFAULT 0 COMMENT '总共消耗的Events数量',
  `includedEvents` int(11) NOT NULL DEFAULT 0 COMMENT '套餐包含的Events数，是否是付费的需要根据paymentlog来定',
  `boughtEvents` int(11) NOT NULL DEFAULT 0 COMMENT '额外购买的Events数',
  `freeEvents` int(11) NOT NULL DEFAULT 0 COMMENT '通过优惠码等方式获得的免费Events',
  `overageEvents` int(11) NOT NULL DEFAULT 0 COMMENT '透支的Events数，展示时计算，账单结算时存库',
  `overageLimit` int(11) NOT NULL DEFAULT 0 COMMENT '能透支的events数量',
  `expired` int(11) NOT NULL DEFAULT 0 COMMENT '0:本期账单;1:以往账单',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserBotBlacklist` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `userId` int(11) NOT NULL,
  `ipRange` text NOT NULL COMMENT 'IpRange字符串数组的json',
  `userAgent` text COMMENT 'UserAgent字符串数组的json',
  `enabled` int(11) NOT NULL DEFAULT 1 COMMENT '0:disabled;1:enabled',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserEventLog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL COMMENT '操作内容所属的User',
  `operatorId` int(11) NOT NULL COMMENT '执行动作的User',
  `operatorIP` varchar(40) NOT NULL DEFAULT '' COMMENT '执行动作的来源IP(ipv4/ipv6)',
  `entityType` int(11) NOT NULL COMMENT '1:Campaign;2:Lander;3:Offer;4:TrafficSource;5:AffiliateNetwork;其他值根据groupby的维度值依次递增',
  `entityTypeString` varchar(64) NOT NULL DEFAULT '' COMMENT 'entityType的明文内容',
  `entityName` text NOT NULL,
  `entityId` varchar(256) NOT NULL,
  `actionType` int(11) NOT NULL COMMENT '1:Create;2:Change;3:Archive;4:Restore;5:Report',
  `changedAt` int(11) NOT NULL COMMENT '创建的时间戳',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Languages` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) NOT NULL,
  `code` varchar(2) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`OSWithVersions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category` varchar(16) NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`BrowserWithVersions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category` varchar(16) NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`BrandWithVersions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category` varchar(16) NOT NULL,
  `name` varchar(128) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`MobileCarriers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `countryCode` varchar(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`TemplatePlan` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `desc` text NOT NULL COMMENT 'HTML格式的描述信息',
  `paypalPlanId` int(11) NOT NULL COMMENT 'PaypalBillingPlan里面的Id',
  `normalPrice` DECIMAL(8,2) NOT NULL COMMENT '原价，仅展示用',
  `onSalePrice` DECIMAL(8,2) NOT NULL COMMENT '促销价格，结算以该价格为准',
  `eventsLimit` int(11) NOT NULL,
  `supportType` varchar(10) NOT NULL COMMENT 'Newbie;Pro;Premium;Dedicated',
  `retentionLimit` int(11) NOT NULL COMMENT '报表最长时间月份数',
  `domainLimit` int(11) NOT NULL COMMENT '可支持Custom Domain的个数:0~',
  `userLimit` int(11) NOT NULL COMMENT '可支持的额外user的个数（不包括自己）:0~',
  `tsReportLimit` int(11) NOT NULL COMMENT '可支持的traffic source report的个数:0~',
  `volumeDiscount` int(11) NOT NULL COMMENT '百分比(50.16%，保存为5016)',
  `overageCPM` int(11) NOT NULL COMMENT '超出流量每千次的价格(0.0231，保存为23100)',
  `overageLimit` int(11) NOT NULL COMMENT '能透支的events数量',
  `regularFrequency` enum('WEEK','DAY','YEAR','MONTH') NOT NULL DEFAULT 'MONTH',
  `regularFrequencyInterval` int(11) NOT NULL DEFAULT '1',
  `hidden` int(11) NOT NULL DEFAULT 0 COMMENT '该Plan是否隐藏起来不让用户看到。0:不隐藏;1:隐藏',
  `order` int(11) NOT NULL DEFAULT 0 COMMENT '排列顺序，数字越小越靠前，0~',
  `orderLimit`  int(11) NOT NULL DEFAULT 0 COMMENT '单用户可以购买几次，0:无限制;>0:最大购买次数',
  `hasCommission` int(11) NOT NULL DEFAULT 0 COMMENT '购买该plan是否引入提成，0:没有;1:有',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除', -- 修改plan时，必须通过新建plan来完成
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Activity` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL COMMENT '运营推广活动的名称',
  `startDay` int(11) NOT NULL COMMENT '可以激活的时间段开始时间点',
  `endDay` int(11) NOT NULL COMMENT '可以激活的时间段结束时间点',
  `userLimit` int(11) NOT NULL DEFAULT 1 COMMENT '该活动的优惠码，每个用户最多可以用几张，0为无限制',
  `couponLimit` int(11) NOT NULL DEFAULT 1 COMMENT '该活动的优惠码，可以同时被多少用户使用，0为无限制',
  `open` int(11) NOT NULL DEFAULT 0 COMMENT '0:关闭;1:开启',
  `desc` text NOT NULL COMMENT 'HTML格式的描述信息',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`Coupon` (
	`id` int(11) NOT NULL AUTO_INCREMENT,
	`code` varchar(32) NOT NULL COMMENT '优惠码',
	`activity` int(11) NOT NULL COMMENT '该优惠码属于哪一次活动',
	`value` int(11) NOT NULL COMMENT '赠送events数量',
	PRIMARY KEY (`id`),
	UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserCouponLog` (
	`id` int(11) NOT NULL AUTO_INCREMENT,
	`activity` int(11) NOT NULL DEFAULT 0 COMMENT '该优惠码属于哪一次活动',
	`couponId` int(11) NOT NULL DEFAULT 0,
	`activateDay` int(11) NOT NULL COMMENT '用户激活优惠码的时间戳',
	`userId` int(11) NOT NULL DEFAULT 0,
	`status` int(11) NOT NULL DEFAULT 0 COMMENT '0:新建;1:已发放;2:已兑现',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- PaypalBillingPlan
CREATE TABLE `PaypalBillingPlan` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `paypalId` varchar(255) NOT NULL COMMENT '在PayPal侧的PlanId',
  `name` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `type` enum('FIXED','INFINITE') NOT NULL DEFAULT 'INFINITE',
  `currency` varchar(3) NOT NULL DEFAULT 'USD',
  `setupFee` decimal(10,2) NOT NULL,
  `regularName` varchar(255) NOT NULL,
  `regularFee` decimal(10,2) NOT NULL,
  `regularCycles` int(11) NOT NULL DEFAULT '0',
  `regularFrequency` enum('WEEK','DAY','YEAR','MONTH') NOT NULL DEFAULT 'MONTH',
  `regularFrequencyInterval` int(11) NOT NULL DEFAULT '1',
  `trialName` varchar(255) NOT NULL,
  `trialFee` decimal(10,2) NOT NULL,
  `trialCycles` int(11) NOT NULL DEFAULT '0',
  `trialFrequency` enum('WEEK','DAY','YEAR','MONTH') NOT NULL DEFAULT 'MONTH',
  `trialFrequencyInterval` int(11) NOT NULL DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `PaypalBillingAgreement` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `paypalPlanId` int(11) NOT NULL,
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0:新建;1:create成功;2:create失败;3:用户同意合约;4:用户拒绝合约;5:解除合约',
  `name` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `token` varchar(255) NOT NULL,
  `approvalUrl` varchar(500) NOT NULL,
  `executeUrl` varchar(500) NOT NULL,
  `createdAt` datetime NOT NULL COMMENT '创建时间',
  `createReq` text COMMENT '创建时发送的request',
  `createResp` text COMMENT '创建时收到的response',
  `cancelReq` text COMMENT '取消合约时发送的request',
  `cancelResp` text COMMENT '取消合约时收到的response',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `PaypalBillingExecute` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `agreementId` int(11) NOT NULL,
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0:execute开始;1:execute成功;2:execute失败;3:已失效',
  `amount` int(11) NOT NULL COMMENT '本次扣款的金额（真实金额*100）',
  `tax` int(11) NOT NULL COMMENT '本次扣款的tax金额（真实金额*100）',
  `executedAt` datetime NOT NULL COMMENT '执行时间',
  `executeReq` text COMMENT '执行时发送的request',
  `executeResp` text COMMENT '执行时收到的response',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- UserBillDetail
CREATE TABLE AdClickTool.`UserBillDetail` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `name` text NOT NULL,
  `address1` text NOT NULL,
  `address2` text NOT NULL,
  `city` text NOT NULL,
  `zip` text NOT NULL,
  `region` text NOT NULL,
  `country` text NOT NULL,
  `taxId` text NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `userId` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- UserPaymentMethod
CREATE TABLE AdClickTool.`UserPaymentMethod` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `type` int(11) NOT NULL COMMENT '支付方式：1(Paypal),2(UnionPay),3(AliPay),4(Master),5(Visa),6(JCB),7(AE),8(WeChat),9(DiscoverCard)...',
  `paypalAgreementId` int(11) DEFAULT 0 NOT NULL COMMENT 'paypal的合约信息Id',
  `info` text NOT NULL COMMENT '各种支付方式不同的数据和格式（付款账号信息等），存储成json',
  `changedAt` int(11) NOT NULL COMMENT '创建的时间戳',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- UserPaymentLog
CREATE TABLE AdClickTool.`UserPaymentLog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `paymentMethod` int(11) NOT NULL COMMENT '支付方式的Id',
  `amount` int(11) NOT NULL COMMENT '本次支付的金额（真实金额*100）',
  `tax` int(11) NOT NULL COMMENT '本次支付的tax金额（真实金额*100）',
  `goodsType` int(11) NOT NULL COMMENT '购买商品的类型(1:Plan;2:Events)',
  `goodsId` int(11) NOT NULL DEFAULT 0 COMMENT '购买商品的Id(PlanId或0)',
  `goodsVolume` int(11) NOT NULL COMMENT '购买商品的数量',
  `timeStamp` int(11) NOT NULL COMMENT '创建的时间戳',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserCommissionLog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `referralId` int(11) NOT NULL COMMENT '用户推荐的记录的Id',
  `paymentLogId` int(11) NOT NULL COMMENT '付款记录的Id',
  `commission` int(11) NOT NULL COMMENT '本次提成的金额（真实金额*1e6）',
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳',
  PRIMARY KEY (`id`),
  UNIQUE KEY `paymentLogId` (`paymentLogId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserEncashLog` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL COMMENT '本次提现佣金的用户Id',
  `amount` int(11) NOT NULL COMMENT '本次提现佣金的金额',
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 角色权限表
CREATE TABLE AdClickTool.`RolePrivilege` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role` int(11) NOT NULL COMMENT '0:group owner;1:group member',
  `config` text NOT NULL COMMENT '角色权限的config信息(json)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 用户功能表
CREATE TABLE AdClickTool.`UserFunctions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `functions` text NOT NULL COMMENT '{TSReport:0~,AdditonalUser:0~,CustomDomain:0~,ANOfferManagement:0~}',
  PRIMARY KEY (`id`),
  UNIQUE KEY `userId` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- 支付宝/微信的二维码支付
CREATE TABLE AdClickTool.`QRPay` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `channel` int(11) NOT NULL DEFAULT 0 COMMENT '支付渠道，0:支付宝;1:微信',
  `tradeNumber` varchar(128) NOT NULL DEFAULT '' COMMENT '支付订单号',
  `goodsType` int(11) NOT NULL COMMENT '购买商品的类型(1:Plan;2:Events)',
  `goodsId` int(11) NOT NULL DEFAULT 0 COMMENT '购买商品的Id(PlanId或0)',
  `goodsVolume` int(11) NOT NULL COMMENT '购买商品的数量',
  `amount` decimal(10,2) NOT NULL COMMENT '支付金额',
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '支付状态,0:新建;1:创建二维码成功;2:创建二维码失败;3:用户支付成功;4:用户支付失败',
  `createdAt` int(11) NOT NULL DEFAULT 0 COMMENT '订单创建的时间戳',
  `createReq` text NOT NULL COMMENT '创建订单时向支付机构发起的请求',
  `createResp` text NOT NULL COMMENT '创建订单时支付机构返回的内容',
  `callback` text NOT NULL COMMENT '订单支付完成时，支付机构回调返回的内容',
  `callbackAt` int(11) NOT NULL DEFAULT 0 COMMENT '订单回调发生的时间戳',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`UserResetCode` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `code` varchar(64) NOT NULL,
  `expireAt` int(11) NOT NULL,
  `status` int(11) NOT NULL DEFAULT 0 COMMENT '0:新建;1:已使用;2:已失效',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`ThirdPartyAffiliateNetwork` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `trustedANId` int(11) NOT NULL,
  `name` varchar(256) NOT NULL DEFAULT '',
  `token` text,
  `userName` text,
  `password` text,
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳，精确到秒',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`OfferSyncTask` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `thirdPartyANId` int(11),
  `status` int(11) DEFAULT 0 COMMENT '0:新建;1:运行中;2:出错;3:完成',
  `executor` varchar(32) DEFAULT '' COMMENT '执行者的唯一标识',
  `message` text COMMENT '出错时的提示消息',
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳，精确到秒',
  `startedAt` int(11) NOT NULL COMMENT '任务开始的时间戳，精确到秒',
  `endedAt` int(11) NOT NULL COMMENT '任务出错或者完成的时间戳，精确到秒',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`ThirdPartyOffer` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `taskId` int(11) NOT NULL,
  `status` int(11) COMMENT '1:active;2:pauseded',
  `offerId` text COMMENT '第三方的OfferId',
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '',
  `previewLink` text,
  `trackingLink` text,
  `countryCode` text COMMENT 'USA,SGP,CHN,IDA,IND',
  `payoutMode` int(11) NOT NULL DEFAULT 1 COMMENT '0:Auto;1:Manual',
  `payoutValue` decimal(10,5) NOT NULL DEFAULT 0,
  `category` text,
  `carrier` mediumtext,
  `platform` text,
  `detail` mediumtext COMMENT 'Offer详细信息的json结构体',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `ThirdPartyTrafficSource` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `trustedTrafficSourceId` int(11) NOT NULL,
  `name` varchar(256) NOT NULL DEFAULT '',
  `token` text,
  `userName` text,
  `password` text,
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳，精确到秒',
  `deleted` int(11) NOT NULL DEFAULT '0' COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `TrafficSourceSyncTask` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `thirdPartyTrafficSourceId` int(11) DEFAULT NULL,
  `status` int(11) DEFAULT '0' COMMENT '0:新建;1:运行中;2:出错;3:检查通过',
  `executor` varchar(32) NOT NULL COMMENT '执行者的唯一标识',
  `message` text COMMENT '出错时的提示消息',
  `tzId` int(11) DEFAULT 35 COMMENT '报表时间段的时区的Id，格式:35',
  `tzShift` varchar(6) DEFAULT '+00:00' COMMENT '报表时间段的时区，格式:+08:00',
  `tzParam` text DEFAULT '' COMMENT '报表时间段的时区，TSService用',
  `statisFrom` datetime NOT NULL COMMENT '报表开始时间',
  `statisTo` datetime NOT NULL COMMENT '报表截止时间',
  `meshSize` int(11) NOT NULL DEFAULT 2 COMMENT '获取报告的粒度，必须要大于该TS能支持的最细粒度，0:minute;1:hour;2:day;3:week;4:month;5:year',
  `createdAt` int(11) NOT NULL COMMENT '创建的时间戳，精确到秒',
  `startedAt` int(11) NOT NULL COMMENT '任务开始的时间戳，精确到秒',
  `endedAt` int(11) NOT NULL COMMENT '任务出错或者完成的时间戳，精确到秒',
  `deleted` int(11) NOT NULL DEFAULT '0' COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `TrafficSourceStatis` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `taskId` int(11) NOT NULL,
  `status` int(11) NOT NULL DEFAULT 1 COMMENT '1:active;2:pauseded',
  `dimensions` int(11) NOT NULL DEFAULT '0' COMMENT '0x1FFF类型的数据(供13bit)，每一个bit标明该行数据是否含有相应的维度，按照v1-v10/campaignId/webSiteId/time的顺序依次赋值',
  `campaignId` varchar(256) NOT NULL DEFAULT '',
  `campaignName` varchar(256) NOT NULL DEFAULT '',
  `websiteId` varchar(256) NOT NULL DEFAULT '',
  `v1` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v2` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v3` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v4` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v5` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v6` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v7` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v8` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v9` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `v10` varchar(255) DEFAULT '' COMMENT '需要根据当前所属的TS的配置来确定具体的token',
  `impression` int(11) DEFAULT 0,
  `click` int(11) DEFAULT 0,
  `cost` bigint(20) NOT NULL DEFAULT '0' COMMENT '实际的值x1000000',
  `time` varchar(20) NOT NULL DEFAULT '' COMMENT '报表记录的时间，格式:2017-04-18 20/2017-04-18/2017-W41/2017-04/2017',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE SCRule2Campaign(
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `ruleId` INT(11),
  `campaignId` INT(11),
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE SuddenChangeRule(
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `userId` INT(11),
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT ''  COMMENT  'rule NAME',
  `dimension` VARCHAR(20) DEFAULT '' COMMENT 'WebSiteId,Country,Carrier,City,Device,OS,OSVersion,ISP,Offer,Lander,Brand,Browser,BrowserVersion',
  `timeSpan` VARCHAR(256) DEFAULT '' COMMENT 'last3hours,last6hours,last12hours,last24hours,last3days,last7days,previousDay,sameDay', 
  `condition` VARCHAR(256) DEFAULT '' COMMENT 'sumImps>500,sumVisits>500,sumClicks<1,ctr<0.5,cr<0.3,cpm>0.02,cpc>0.5,cpa>0.1',
  `schedule` VARCHAR(16) DEFAULT '' COMMENT '记录定期的任务描述：0 0 * * * *',
  `oneTime` VARCHAR(16) DEFAULT '' COMMENT '记录仅执行一次的任务的描述：2017-04-21 23',
  `scheduleString` VARCHAR(128) DEFAULT '' COMMENT 'Every 1/3/6/12 Hour,Daily 23,Weekly 0 23,One Time 2017-04-21 23',
  `emails` TEXT COMMENT 'hemin@adbund.com,hemin@innotechx.com',
  `status` INT(2) NOT NULL DEFAULT '0' COMMENT 'active: 1,inactive: 0',
  `deleted` int(11) NOT NULL DEFAULT '0' COMMENT '0:未删除;1:已删除',
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE SuddenChangeLog(
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `ruleId` INT(11) UNSIGNED NOT NULL,
  `dimension` VARCHAR(20) DEFAULT '' COMMENT 'WebSiteId,Country,Carrier,City,Device,OS,OSVersion,ISP,Offer,Lander,Brand,Browser,BrowserVersion',
  `hit` INT(2) NOT NULL DEFAULT 0 COMMENT '0:未命中,1:命中',
  `condition` VARCHAR(256) DEFAULT '' COMMENT 'sumImps>500,sumVisits>500,sumClicks<1,ctr<0.5,cr<0.3,cpm>0.02,cpc>0.5,cpa>0.1',
  `notifiedEmails` TEXT ,
  `sendStatus` INT(2) DEFAULT '0' ,
  `timeStamp` INT(10) DEFAULT NULL COMMENT 'unix时间戳',
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE SuddenChangeLogDetail(
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `logId` INT(11) UNSIGNED NOT NULL,
  `campaignID` INT(11),
  `data` text NOT NULL  COMMENT '命中情况下，存储符合rule的维度数据的数组(json)',
  `dimensionKey` varchar(20) DEFAULT NULL,
  `dimensionValue` varchar(2000) DEFAULT NULL,
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE FFRule2Campaign(
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `ruleId` INT(11),
  `campaignId` INT(11),
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE FraudFilterRule(
  `id` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `userId` INT(11),
  `name` varchar(256) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL DEFAULT '' COMMENT 'rule NAME',
  `dimension` VARCHAR(20) DEFAULT '' COMMENT 'IP',
  `timeSpan` INT(11) DEFAULT 0 COMMENT '单位：秒', 
  `condition` VARCHAR(256) DEFAULT '' COMMENT 'PV>500,UserAgent>100,Clicks>100',
  `status` INT(2) NOT NULL DEFAULT '0' COMMENT 'active: 1,inactive: 0',
  `deleted` int(11) NOT NULL DEFAULT '0' COMMENT '0:未删除;1:已删除',
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE FraudFilterLog(
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `ruleId` INT(11) UNSIGNED NOT NULL,
  `dimension` VARCHAR(20) DEFAULT '' COMMENT 'IP',
  `hit` INT(2) NOT NULL DEFAULT 0 COMMENT '0:未命中,1:命中',
  `condition` VARCHAR(256) DEFAULT '' COMMENT 'PV>500,UserAgent>100,Clicks>100',
  `timeStamp` INT(10) DEFAULT NULL COMMENT 'unix时间戳',
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE FraudFilterLogDetail(
  `id` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `logId` INT(11) UNSIGNED NOT NULL,
  `campaignID` INT(11),
  `data` text NOT NULL  COMMENT '命中情况下，存储符合rule的维度数据的数组(json)',
 PRIMARY KEY (`id`)
) ENGINE=INNODB DEFAULT CHARSET=utf8;

CREATE TABLE `ISP` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(256) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
