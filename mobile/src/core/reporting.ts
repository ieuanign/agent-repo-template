// The one place a caught error leaves the app; an error tracker's capture call goes here.
export function reportError(error: Error): void {
  console.error(error);
}
