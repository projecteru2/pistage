/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pistage_run_tab` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `create_time` bigint(20) unsigned NOT NULL,
  `update_time` bigint(20) unsigned NOT NULL,
  `start_time` bigint(20) unsigned NOT NULL,
  `end_time` bigint(20) unsigned NOT NULL,
  `workflow_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `workflow_identifier` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `snapshot_version` bigint(20) unsigned NOT NULL,
  `run_status` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  KEY `uk_run` (`workflow_identifier`,`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `pistage_snapshot_tab` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `create_time` bigint(20) unsigned NOT NULL,
  `update_time` bigint(20) unsigned NOT NULL,
  `workflow_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `workflow_identifier` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `content` text COLLATE utf8mb4_unicode_ci NOT NULL,
  `content_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_snapshot` (`workflow_identifier`,`content_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
CREATE TABLE `job_run_tab` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `create_time` bigint(20) unsigned NOT NULL,
  `update_time` bigint(20) unsigned NOT NULL,
  `start_time` bigint(20) unsigned NOT NULL,
  `end_time` bigint(20) unsigned NOT NULL,
  `workflow_type` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `workflow_identifier` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `pistage_run_id` bigint(20) unsigned NOT NULL,
  `job_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  `run_status` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_run` (`pistage_run_id`, `job_name`),
  KEY `idx_job_run` (`workflow_identifier`, `job_name`, `create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
