export function snakeToSentence(string: string) {
    if (!string) {
        return "";
    }
    // If key contains dots, use last segment
    const k = string.includes(".") ? string.split(".").pop()! : string;
    // replace underscores and dashes with spaces, separate camelCase
    const withSpaces = k
        .replace(/[_-]+/g, " ")
        .replace(/([a-z0-9])([A-Z])/g, "$1 $2");
    // collapse spaces, trim
    const cleaned = withSpaces.replace(/\s+/g, " ").trim();
    // capitalize first letter
    return cleaned.charAt(0).toUpperCase() + cleaned.slice(1);
}

export function snakeToTitle(string: string) {
    if (!string) {
        return "";
    }

    return snakeToSentence(string)
        .split(" ")
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(" ");
}
