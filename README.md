# Weditor - Modern Markdown Editor

Weditor is a modern Markdown editor built on the Wails framework. It offers rich editing features and an elegant user interface, with support for live preview, WYSIWYG editing, theme customization, and more.

## 🚀 Key Features

### Editing
- **Live preview**: See the rendered result in real time as you edit Markdown content
- **WYSIWYG**: Visual editing mode with no need to memorize Markdown syntax
- **Markdown syntax guide**: Built-in syntax reference to help you master Markdown writing quickly
- **Code highlighting**: Syntax highlighting for code blocks in multiple programming languages
- **Mermaid diagram support**: Create and render Mermaid diagrams directly in the editor
- **Rich toolbar**: Shortcut buttons for common editing operations
  - **Headings**: Heading 1, Heading 2, Heading 3
  - **Text formatting**: Bold, italic, strikethrough
  - **Links and images**: Insert link, insert image
  - **Code**: Inline code, code block
  - **Lists and quotes**: Blockquote, unordered list, ordered list, insert table
  - **Other**: Horizontal rule
- **Scroll sync**: Automatic scrolling sync between the editor and preview panes
- **Real-time content updates**: Edits are saved and updated in real time

### Themes & Styling
- **Multiple themes**: Built-in color themes with light/dark mode switching
- **Custom themes**: Customize the editor's color scheme
- **Font settings**: Adjust font family and size
- **Code styles**: Choose from different code highlighting styles
- **Spacing adjustments**: Customize line spacing and paragraph spacing
- **Theme system**: Support for different theme systems (e.g., Breeze and Default)

### User Experience
- **Responsive design**: Adapts to different screen sizes
- **Back-to-top button**: Quickly return to the top when editing long documents
- **Notification system**: Action feedback and alerts
- **i18n support**: Switch between Chinese and English
- **Clipboard management**: Enhanced clipboard features
- **Auto-update**: Checks for new versions after startup; download, replacement and restart in one click (macOS / Windows), with stable and development channels

### Editor Features
- **Toolbar**: Shortcut buttons for common editing operations
- **Settings panel**: Rich configuration options to customize editor behavior
- **Floating settings panel**: Tabbed settings interface
- **Color picker**: Intuitive color customization tool
- **Dropdown menus**: Convenient access to features

## 🛠️ Tech Stack

### Backend
- **Go**: High-performance backend development
- **Wails framework**: Cross-platform desktop application framework

### Frontend
- **Vue.js 3**: Modern frontend framework
- **Vite**: Fast build tool
- **Composition API**: More flexible organization of component logic
- **CSS variables**: Dynamic theme customization

### Core Libraries
- **Prism.js**: Code highlighting
- **Mermaid**: Diagram rendering
- **Custom Markdown parser**: Optimized Markdown processing system

## 📦 Installation & Running

### Prerequisites
- Go 1.18+ (for backend development)
- Node.js 16+ (for frontend development)
- Wails CLI (for building the project)

### Installation Steps

1. **Clone the project**
   ```bash
   git clone <repository-url>
   cd Weditor
   ```

2. **Install dependencies**
   ```bash
   # Install frontend dependencies
   cd frontend
   npm install
   cd ..
   
   # Install backend dependencies
   go mod tidy
   ```

3. **Run in development mode**
   ```bash
   wails dev
   ```

4. **Build the production version**
   ```bash
   wails build
   ```

## 📖 Usage Guide

### Basic Editing
1. Type Markdown content in the editor pane on the left
2. The rendered result is shown in real time on the right
3. Switch to WYSIWYG mode for visual editing

### Theme Customization
1. Click the settings button in the top-right corner to open the settings panel
2. Select the "Color Theme" tab
3. Choose a preset theme or customize the colors
4. Adjust fonts, code styles, and spacing

### Using Mermaid Diagrams
1. Enter a Mermaid code block in the editor
   ```mermaid
   graph TD;
       A-->B;
       A-->C;
       B-->D;
       C-->D;
   ```
2. An interactive diagram will appear after rendering


### Code Highlighting
1. Enter a code block with a language identifier in the editor
   ```javascript
   function hello() {
       console.log("Hello, Weditor!");
   }
   ```
2. Syntax highlighting for the corresponding language will appear after rendering

## 🔄 Releases & Auto-Update

### Publishing a new version

```bash
git tag v1.1.0 && git push origin v1.1.0
```

CI builds all three platforms in parallel and additionally generates a `latest.json` update manifest (download URL and sha256 for every platform's artifact), which is uploaded alongside the Release. Installed clients detect the new version and prompt on their next launch.

Pushing to `main` / `master` / `release` instead publishes a **pre-release** tagged `v1.0.0-dev-<sha>`; only clients switched to the "Development" channel will see it.

### How auto-update works

| Stage | Implementation |
|---|---|
| Version source | The git tag is injected at build time via `-ldflags -X`, and synchronised into the macOS plist and the Windows version resource |
| Discovery | The stable channel reads `releases/latest/download/latest.json` (a static URL that costs no GitHub API quota); the development channel asks the Releases API for the most recently published build |
| Newer or older | The stable channel compares semver; the development channel compares publish times, because dev versions look like `v1.0.0-dev-<sha>` and cannot be ordered by version number |
| Integrity | sha256 is computed while streaming the download and compared with the manifest — a mismatch is discarded at once, and the file is verified again right before installing |
| macOS install | `ditto` extracts the archive; a detached script then waits for the app to exit, swaps in the whole `.app` bundle and relaunches it |
| Windows install | The running exe is renamed out of the way, the new version takes its place, and the new process is started |
| Trigger | A quiet check five seconds after startup; the Help ▸ "Check for Updates…" menu item and the Updates tab in Settings trigger it manually |

The Settings ▸ Updates tab lets you toggle automatic checks, switch between the stable and development channels, and skip a specific version.

> These artifacts are neither code-signed nor notarised: on macOS you still have to allow the app manually the first time, and Windows SmartScreen may warn about an unknown publisher on first run.

## ⚙️ Configuration Options

### Editor Settings
- **Font family**: Choose the font used by the editor
- **Font size**: Adjust the font size in the editing area
- **Line spacing**: Set the spacing between lines
- **Paragraph spacing**: Set the spacing between paragraphs

### Theme Settings
- **Color theme**: Choose a preset theme or customize colors
- **Code style**: Choose a code highlighting style
- **Theme system**: Switch between different theme systems

### Interface Settings
- **Language**: Switch between Chinese and English
- **Notifications**: Turn notifications on or off

## 🏗️ Project Structure

```
Weditor/
├── build/              # Build configuration and assets
├── frontend/           # Frontend code
│   ├── src/            # Source code
│   │   ├── components/ # Vue components
│   │   ├── composables/ # Composition functions
│   │   ├── config/     # Configuration files
│   │   ├── core/       # Core feature implementations
│   │   ├── locales/    # i18n resources
│   │   ├── plugins/    # Plugins
│   │   ├── shared/     # Shared utilities
│   │   └── styles/     # Style files
│   └── wailsjs/        # Wails runtime bindings
├── app.go              # Backend application logic
├── main.go             # Application entry point
├── version.go          # Version parsing and comparison
├── updater.go          # Auto-update: manifest fetch / download / verification / orchestration
├── updater_apply_*.go  # Per-platform install (darwin / windows / other)
└── wails.json          # Wails configuration
```

## 🤝 Development Notes

### Frontend Development
```bash
cd frontend
npm run dev
```

### Backend Development
```bash
# Run backend tests
go test ./...
```

### Code Style
- Frontend: Use ESLint and Prettier to keep code style consistent
- Backend: Follow Go's standard code style

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Issues and Pull Requests are welcome to help improve Weditor!

## 📞 Contact

If you have questions or suggestions, feel free to reach out:

- Submit an Issue: [GitHub Issues](<repository-issues-url>)
- Send an email: <contact-email>

---

**Weditor** - Making Markdown editing more efficient and elegant!
