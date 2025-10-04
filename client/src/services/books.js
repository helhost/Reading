
class BookService {
  async getAll() {
    try {
      const res = await fetch("http://localhost:8080/books");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json()
    } catch (err) {
      console.error("Request failed:", err);
    }
  }

}

let bookServcie = new BookService()
export default bookServcie
