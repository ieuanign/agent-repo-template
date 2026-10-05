// A literal dot read: Metro inlines only that form into the bundle.
const apiUrlValue = process.env.EXPO_PUBLIC_API_URL;
if (!apiUrlValue) {
  throw new Error("EXPO_PUBLIC_API_URL is empty or unset.");
}
export const apiUrl: string = apiUrlValue;
