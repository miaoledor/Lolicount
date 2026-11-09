// Build-time resolution of the version string shown in the footer. The value
// is baked into the SSG output, so releasing a new tag never requires editing
// a source file: CD passes the tag as a Docker build arg, GitHub tag builds
// expose GITHUB_REF_NAME, and local builds read the nearest reachable tag.
import { execFileSync } from 'node:child_process';

export const VERSION_TAG_RE = /^v\d+\.\d+\.\d+/;

// describeNearestTag stays silent when git metadata is unavailable (for
// example inside the Docker build context, where .git is excluded).
export const describeNearestTag = () => {
  try {
    return execFileSync('git', ['describe', '--tags', '--abbrev=0'], {
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    }).trim();
  } catch {
    return '';
  }
};

// resolveAppVersion prefers an explicit CD/build value, then a GitHub tag
// ref, then the nearest local tag. An empty result means the build has no
// trustworthy version and the footer simply omits it. env and describeTag
// are injectable so the precedence can be unit tested without git.
export const resolveAppVersion = (env = process.env, describeTag = describeNearestTag) => {
  const explicit = (env.NUXT_PUBLIC_APP_VERSION || env.APP_VERSION || '').trim();
  if (explicit) return explicit;
  const refName = (env.GITHUB_REF_NAME || '').trim();
  if (VERSION_TAG_RE.test(refName)) return refName;
  const tag = (describeTag() || '').trim();
  return VERSION_TAG_RE.test(tag) ? tag : '';
};
