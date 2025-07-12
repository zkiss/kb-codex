import React from "react";
import { marked } from "marked";

export default function MarkdownViewer({ content }) {
  return (
    <div
      className="mt-2"
      dangerouslySetInnerHTML={{ __html: marked.parse(content) }}
    ></div>
  );
} 