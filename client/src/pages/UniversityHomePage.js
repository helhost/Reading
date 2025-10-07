import getMe from "../util/index.js";

export default async function UniversityHomePage(id) {
  const me = await getMe();
  if (!me) return;

  console.log(`a page where you can see, join and leave courses for university: ${id}`)
}
