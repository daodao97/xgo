/*
 Navicat Premium Data Transfer

 Source Server         : claude_flex
 Source Server Type    : MySQL
 Source Server Version : 80025 (8.0.25)
 Source Host           : rm-wz90p4o62oa0l7r6gao.mysql.rds.aliyuncs.com:3306
 Source Schema         : claude_flex

 Target Server Type    : MySQL
 Target Server Version : 80025 (8.0.25)
 File Encoding         : 65001

 Date: 15/08/2024 11:39:36
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for operator
-- ----------------------------
DROP TABLE IF EXISTS `operator`;
CREATE TABLE `operator` (
                            `id` int NOT NULL AUTO_INCREMENT,
                            `username` varchar(255) NOT NULL DEFAULT '',
                            `password` varchar(255) NOT NULL DEFAULT '',
                            `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                            PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of operator
-- ----------------------------
BEGIN;
INSERT INTO `operator` (`id`, `username`, `password`, `created_at`, `updated_at`) VALUES (1, 'test', '$2a$10$ylNOtEixT/OsUja.C2mI..EMYzCEW6V5qk/heBfH65lZTFX7A/FNe', '2024-07-15 06:24:57', '2024-07-15 07:45:00');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
