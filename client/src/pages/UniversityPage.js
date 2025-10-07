import getMe from "../util/index.js";

export default async function UniversityPage() {
  const me = await getMe();
  if (!me) return;

  console.log("page for joining and leaving universiteis")
}
