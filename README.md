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
