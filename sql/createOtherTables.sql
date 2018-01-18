CREATE TABLE AdClickTool.`CustomerContact` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `websiteUrl` varchar(250) NOT NULL DEFAULT '',
  `email` varchar(100) NOT NULL,
  `tel` varchar(50) NOT NULL DEFAULT '',
  `intention` varchar(256) NOT NULL DEFAULT '',
  `message` text NOT NULL DEFAULT '',
  `createdAt` int(11) NOT NULL DEFAULT 0 COMMENT '创建的时间戳',
  `deleted` int(11) NOT NULL DEFAULT 0 COMMENT '0:未删除;1:已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;

CREATE TABLE AdClickTool.`CampaignMap` (
  `OurCampId` int(11) NOT NULL COMMENT '我们的CampaignId',
  `Timestamp` bigint(22) NOT NULL COMMENT '精确到15分钟的时间戳(如：1486698300000)',
  `TheirCampId` varchar(128) NOT NULL COMMENT '他们的CampaignId',
  PRIMARY KEY (`OurCampId`,`Timestamp`,`TheirCampId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE Cache.`ReqCache` (
  `clickId` varchar(250) NOT NULL,
  `value` text NOT NULL,
  PRIMARY KEY (`clickId`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8;