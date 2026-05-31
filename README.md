# CaspawCG

A GoLang + Wails application designed to integrate the CasparCG server with dynamic data sources.

CaspawCG enables real-time graphics control for live broadcasting by connecting CasparCG Server with external data sources like Google Sheets, providing a UI inspired by CharacterWorks.

- [CaspawCG](#caspawcg)
  - [Features](#features)
- [For Developers](#for-developers)
  - [Prerequisites](#prerequisites)
  - [Development](#development)
  - [Configuration](#configuration)
  - [How to get Google `credentials.json`?](#how-to-get-google-credentialsjson)
  - [Contributing](#contributing)
  - [License](#license)
  - [Acknowledgments](#acknowledgments)
  - [Support](#support)

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
git clone https://github.com/overlayfox/CaspawCG.git
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

# Install CasparCG Server on Ubuntu24.04
sudo apt update

wget https://github.com/CasparCG/server/releases/download/v2.5.0-stable/casparcg-cef-142_142.0.17.g60aac24+2-noble1_amd64.deb
sudo apt install ./casparcg-cef-142_142.0.17.g60aac24+2-noble1_amd64.deb

wget https://github.com/CasparCG/server/releases/download/v2.5.0-stable/casparcg-server-2.5_2.5.0.stable-noble1_amd64.deb
sudo apt install ./casparcg-server-2.5_2.5.0.stable-noble1_amd64.deb

# Install CasparCG Media Server on Ubuntu 24.04
wget https://github.com/CasparCG/media-scanner/releases/download/v1.4.0/casparcg-scanner_1.4.0-ubuntu1_amd64.deb
sudo apt install ./casparcg-scanner_1.4.0-ubuntu1_amd64.deb
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

## How to get Google `credentials.json`?

1. Go to [Googles Cloud Console](https://console.cloud.google.com).
2. Create a new project that fits your requirements.
3. Go to [Google Sheet API](https://console.cloud.google.com/marketplace/product/google/sheets.googleapis.com).
4. Enable the Google Sheets API -> then click `Manage`.
5. Now select `Credentials` on the left hand side.
6. Click `Create credentials` -> select `Service Account`
7. Fill in the required fields and click `Create`, simply skip the permissions step
8. Now download the JSON and rename it to `credentials.json`
9. Drop it into the project and link to it in the `config.yaml` under `data_source_manager:` -> `google_sheet_data_sources:` --> `credentials_file_path:`
10. Now invite the E-Mail address that is listed in `credentials.json` -> `client_email:` to your Google Sheet and give it viewing access
11. Finally copy the google sheets `spreadsheet_id`, which is located in the URL of your google sheets: `https://docs.google.com/spreadsheets/d/THIS_HERE_IS_YOUR_SPREAD_SHEET_ID/`

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
