# API Model Forge v1.0.1

- Rejects malformed trailing JSON and duplicate JSON keys before generation.
- Allocates deterministic, unique identifiers when property names collide, and avoids generated type/import conflicts.
- Escapes quotes, backslashes and control characters in generated names, attributes and comments. Go preserves unusual wire names with explicit JSON methods.
- Generates valid Kotlin classes for empty objects and safe JVM property names; renamed Kotlin and Python fields retain original JSON names in comments.
- Clears stale output when input is empty or cleared.
- Adds a distinct structured-data Windows application icon.
- Keeps JSON processing fully offline.

Microsoft Edge WebView2 Runtime is required separately. The portable package does not bundle or automatically install it; the missing-runtime dialog offers Microsoft's official download page.
