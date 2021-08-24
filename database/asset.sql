CREATE DATABASE cdr;

\c cdr

CREATE TABLE cdr (
  ANUM VARCHAR(255),
  BNUM VARCHAR(255),
  ServiceType VARCHAR(255),
  CallCategory VARCHAR(255),
  SubscriberType VARCHAR(255),
  StartDatetime timestamp default NULL,
  UsedAmount VARCHAR(255),
  RoundedUsedAmount VARCHAR(255),
  Charge numeric(10,5),
  VoiceCharge numeric(10,5),
  GprsCharge numeric(10,5),
  SmsCharge numeric(10,5)
);
