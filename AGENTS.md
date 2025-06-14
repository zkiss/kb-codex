# Application Description

We are building an application that enables users to create knowledge bases,
upload documents and ask questions about them.

# Initial features

- Authentication with simple email/password
- Create knowledge bases
- Upload text files (txt, md) into knowledge bases
- Ask questions about the knowledge base via chat interface

# Technical notes

- Go backend
- Web frontend
- Use RAG to answer the user questions
    - connect to postgresql db with pgvector plugin (see run-db.sh)
    - schema DDL to be in sql files, using sequential migration files, to be applied on backend startup sequence
    - store document chunks in vector column
- Use websockets as appropriate by functionality
- Prefer using existing opensource libraries instead of reinventing the wheel
- Follow best practices and idiomatic ways to use libraries

# Target features

- Support Auth with popular providers (Google, Facebook, etc)
- Support more media types in knowledge base: pdf, audio, video, spreadsheet, etc
- Full management functionality for files and knowledge bases (create, delete, rename, etc)
- Support voice QnA
- Support updating knowledge bases via chat/voice, as appropriate based on the conversation
    - record the evolution of facts and updates
    - ability to detect when piece of certain knowledge is superseded by a new one, always use latest
