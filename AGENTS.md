<!-- BEGIN:react-hooks-rules -->
## React: useEffect dependencies
Always include ALL variables referenced inside `useEffect` in the dependency array — no more, no less. Using a variable in the callback (e.g. in a condition like `!accountId`) without listing it as a dependency is a bug.
<!-- END:react-hooks-rules -->

<!-- BEGIN:nextjs-agent-rules -->
# This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` before writing any code. Heed deprecation notices.
<!-- END:nextjs-agent-rules -->
