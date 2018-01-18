DELIMITER $$

DROP PROCEDURE IF EXISTS `AdClickTool`.`setPlan` $$
CREATE PROCEDURE `AdClickTool`.`setPlan` (IN changeUserId INT, IN templatePlanId INT, IN moneyAmount INT)
BEGIN

  DECLARE planEventsLimit INT;
  DECLARE planDomainLimit INT;
  DECLARE planUserLimit INT;
  DECLARE planTSReportLimit INT;
  DECLARE planRetentionLimit INT;
  
  SET planEventsLimit = 0;
  
  SELECT 
  `eventsLimit`,`domainLimit`,`userLimit`,`tsReportLimit`,`retentionLimit` 
  INTO planEventsLimit,planDomainLimit,planUserLimit,planTSReportLimit,planRetentionLimit 
  FROM `TemplatePlan` where id =  templatePlanId;
  
  UPDATE `User` SET `status` = 1 where `id` = changeUserId;
 
  UPDATE `UserBilling` SET `expired` = 1 WHERE `userId` = changeUserId;
 
  INSERT INTO `UserBilling` 
    (`userId`, `agreementId`, `planId`, `planPaymentLogId`, `billedEvents`, `totalEvents`, `freeEvents`, `overageEvents`, `planStart`, `planEnd`, `includedEvents`, `nextPlanId`, `nextPaymentMethod`, `expired`)
  VALUES
    (changeUserId, 0, templatePlanId, 0, 0, 0, 0, 0, UNIX_TIMESTAMP(NOW()), UNIX_TIMESTAMP(DATE_ADD(NOW(), INTERVAL 1 MONTH)), planEventsLimit, 0, 0, 0);

  SET @functions = CONCAT('{"domainLimit":',planDomainLimit,',"userLimit":',planUserLimit,',"tsReportLimit":',planTSReportLimit,',"retentionLimit":',planRetentionLimit,',"retentionLimit":',planRetentionLimit,'}');
  INSERT INTO `UserFunctions`(`userId`,`functions`) VALUES(changeUserId,@functions) ON DUPLICATE KEY UPDATE `functions`=@functions;

  INSERT INTO `UserPaymentLog`
	(`userId`,`amount`,`goodsType`,`goodsId`,`goodsVolume`,`timeStamp`)
  VALUES
	(changeUserId,moneyAmount,1,templatePlanId,1,UNIX_TIMESTAMP());

  SELECT 'User Plan Added' AS Message, changeUserId, templatePlanId, moneyAmount;
END $$

DELIMITER ;

DELIMITER $$

DROP PROCEDURE IF EXISTS `AdClickTool`.`setPlanEmail` $$
CREATE PROCEDURE `AdClickTool`.`setPlanEmail` (IN userEmail VARCHAR(100), IN templatePlanId INT, IN moneyAmount INT)
BEGIN

  DECLARE changeUserId INT;
  SET changeUserId = 0;
  SELECT `id` INTO changeUserId
  FROM `User` WHERE `email` = userEmail;
  
  IF changeUserId > 0 THEN
	CALL setPlan(changeUserId, templatePlanId, moneyAmount);
  ELSE
	SELECT 'User Not Exist' AS Message,userEmail,templatePlanId, moneyAmount;
  END IF;

END $$

DELIMITER ;

-- echo "PUBLISH channel_campaign_changed_users 0.update.user.8" | redis-cli -h ec2-52-14-89-249.us-east-2.compute.amazonaws.com -a R%LKsIJF412 --pipe