/** contentPath maps a polymorphic content reference (content_type + id) to its
 * detail route. Used by track programs, notes and bookmarks to link back to the
 * referenced content. Article/problem detail pages accept the id (backend D-026
 * resolves id-or-slug); glossary has no per-term page, so it links to the list. */
export function contentPath(type: string, id: string): string {
  switch (type) {
    case "video":
      return `/videos/${id}`
    case "article":
      return `/articles/${id}`
    case "quiz":
      return `/quizzes/${id}`
    case "problem":
      return `/problems/${id}`
    case "project":
      return `/projects/${id}`
    case "track":
      return `/tracks/${id}`
    case "cheatsheet":
      return `/cheatsheets/${id}`
    case "glossary":
      return `/glossary`
    default:
      return "/"
  }
}
