CREATE TABLE `@db`.`@table` (
  `ID` BIGINT(10) NOT NULL COMMENT '自增主键',
  `create_at` DATETIME NOT NULL DEFAULT now() COMMENT '创建时间',
  `action` VARCHAR(50) NOT NULL COMMENT '动作，支持[kill , freeze]',
  `kind` VARCHAR(100) NOT NULL COMMENT '资源类型，比如deployment, pod，statefulset 等',
  `name` VARCHAR(200) NOT NULL COMMENT '资源名称',
  `expect_execute_at` DATETIME NOT NULL COMMENT '期望执行时间',
  `actual_execute_at` DATETIME NULL COMMENT '实际执行时间',
  `result` SMALLINT(2) NOT NULL DEFAULT 0 COMMENT '实际执行结果。0：初始状态，尚未执行；1：执行成功；2：执行失败；3：等待下一个执行周期。',
  `job_type` VARCHAR(45) NOT NULL DEFAULT 'once' COMMENT '支持 once, period',
  `cron_expression` VARCHAR(45) NOT NULL DEFAULT '‘’' COMMENT '预留字段，暂未实现',
  PRIMARY KEY (`ID`));