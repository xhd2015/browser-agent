/**
 * Canonical browser-agent job type string constants.
 * Shared protocol surface with Go (browseragent.JobType* / IsKnownJobType).
 */

export const JOB_TYPE_INFO = "info";
export const JOB_TYPE_EVAL = "eval";
export const JOB_TYPE_RUN = "run";
export const JOB_TYPE_LOGS = "logs";
export const JOB_TYPE_SCREENSHOT = "screenshot";
export const JOB_TYPE_CDP = "cdp";

/** All known job types in stable order. */
export const KNOWN_JOB_TYPES = [
  JOB_TYPE_INFO,
  JOB_TYPE_EVAL,
  JOB_TYPE_RUN,
  JOB_TYPE_LOGS,
  JOB_TYPE_SCREENSHOT,
  JOB_TYPE_CDP,
] as const;

export type JobType = (typeof KNOWN_JOB_TYPES)[number];

export function isKnownJobType(s: string): s is JobType {
  return (KNOWN_JOB_TYPES as readonly string[]).includes(s);
}
