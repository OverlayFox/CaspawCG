# CaspawCG

A GoLang + Wails application designed to integrate the CasparCG server with dynamic data sources.

CaspawCG enables real-time graphics control for live broadcasting by connecting CasparCG Server with external data sources like Google Sheets, providing a UI inspired by CharacterWorks.

## Features

- 🎨 **Live Graphics Control** - Real-time control of CasparCG graphics and templates
- 📊 **Data Source Integration** - Pull data from Google Sheets and other sources
- 🖥️ **Modern UI** - Built with Wails for a native desktop experience
- ⚡ **Real-time Updates** - Event-driven architecture for instant data synchronization
- 🎯 **Layout Management** - Flexible layout configuration system
- 🔧 **Multi-Server Support** - Connect to multiple CasparCG servers simultaneously

---

# For Developers

## Prerequisites

- [Go](https://golang.org/dl/) 1.26 or higher
- [Node.js](https://nodejs.org/) 16+ and npm
- [Wails](https://wails.io/) v2
- [CasparCG Server](https://github.com/CasparCG/server)
- [CasparCG Media Scanner](https://github.com/CasparCG/media-scanner)

## Development

1. Clone the repository:

```bash
git clone https://github.com/yourusername/CaspawCG.git
cd CaspawCG
```

2. Install dependencies:

```bash
# Install Go dependencies
go mod download

# Install Linter
make install-dev

# Install frontend dependencies
cd frontend
npm install
cd ..
```

3. Start CasparCG

```bash
# Starts CasparCG 2.5.0 with a base config
# Run each make command in a new terminal session
make caspar-scanner
make caspar-server
```

4. Run the application:

```bash
# Run in development mode
make dev
```

## Configuration

1. Copy the example configuration file:

```bash
cp config.example.yaml config.yaml
```

2. Edit `config.yaml` with your settings:

```yaml
data_source_manager:
  google_sheet_data_sources:
    - spreadsheet_id: "your_spreadsheet_id_here"
      credentials_file_path: "path/to/credentials.json"

casparcg_client:
  - host: "localhost"
    port: 5250
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [CasparCG Server](https://github.com/CasparCG/server) - Open-source graphics server
- [Wails](https://wails.io/) - Go + Web UI framework
- [Google Sheets API](https://developers.google.com/sheets/api) - Data source integration

## Support

For issues, questions, or suggestions:

- Open an issue on GitHub
- Check the [CasparCG documentation](http://casparcg.com/wiki/)
- Visit the [Wails documentation](https://wails.io/docs/introduction)
