CREATE TABLE `@db`.`@table` (
  `ID` BIGINT(10) NOT NULL COMMENT 'auto increment ID',
  `create_at` DATETIME NOT NULL DEFAULT now() ,
  `action` VARCHAR(50) NOT NULL COMMENT 'support value: kill , freeze',
  `kind` VARCHAR(100) NOT NULL COMMENT 'resource type ,support value: deployment, pod，statefulset 等',
  `name` VARCHAR(200) NOT NULL COMMENT 'resource name',
  `expect_execute_at` DATETIME NOT NULL ,
  `actual_execute_at` DATETIME NULL ,
  `result` SMALLINT(2) NOT NULL DEFAULT 0 COMMENT '0：default status ；1：success ；2：fail；3：waiting for next schedule',
  `job_type` VARCHAR(45) NOT NULL DEFAULT 'once' COMMENT 'support value: once, period',
  `cron_expression` VARCHAR(45) NOT NULL DEFAULT '‘’' COMMENT 'unused field',
  PRIMARY KEY (`ID`));