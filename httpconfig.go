package talkkonnect

import (
	//	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	// "os"
)

const configTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Talkkonnect Configuration Editor</title>
    <style>
        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #0f172a;
            color: #e2e8f0;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }

        .container {
            width: 90%;
            max-width: 1000px;
            background: rgba(30, 41, 59, 0.7);
            backdrop-filter: blur(10px);
            padding: 30px;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
            border: 1px solid #334155;
            display: flex;
            flex-direction: column;
            gap: 20px;
            height: 90vh;
            box-sizing: border-box;
        }

        h1 {
            margin: 0;
            font-size: 24px;
            font-weight: 600;
            color: #f8fafc;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .status {
            font-size: 14px;
            padding: 8px 16px;
            border-radius: 6px;
            font-weight: 500;
        }

        .status.success {
            background-color: rgba(16, 185, 129, 0.2);
            color: #10b981;
            border: 1px solid rgba(16, 185, 129, 0.3);
        }

        .status.error {
            background-color: rgba(239, 68, 68, 0.2);
            color: #ef4444;
            border: 1px solid rgba(239, 68, 68, 0.3);
        }

        form {
            display: flex;
            flex-direction: column;
            flex-grow: 1;
            gap: 15px;
        }

        textarea {
            flex-grow: 1;
            width: 100%;
            padding: 20px;
            font-family: 'Fira Code', 'Courier New', Courier, monospace;
            font-size: 14px;
            line-height: 1.5;
            border-radius: 8px;
            border: 1px solid #475569;
            background-color: #1e293b;
            color: #cbd5e1;
            resize: none;
            box-sizing: border-box;
            outline: none;
            transition: border-color 0.2s, box-shadow 0.2s;
        }

        textarea:focus {
            border-color: #3b82f6;
            box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.2);
        }

        .actions {
            display: flex;
            justify-content: flex-end;
            align-items: center;
        }

        button {
            padding: 12px 24px;
            font-size: 16px;
            font-weight: 600;
            color: white;
            background: linear-gradient(135deg, #3b82f6, #2563eb);
            border: none;
            border-radius: 6px;
            cursor: pointer;
            transition: transform 0.1s, box-shadow 0.2s;
            box-shadow: 0 4px 6px rgba(37, 99, 235, 0.3);
        }

        button:hover {
            transform: translateY(-1px);
            box-shadow: 0 6px 10px rgba(37, 99, 235, 0.4);
        }

        button:active {
            transform: translateY(1px);
            box-shadow: 0 2px 4px rgba(37, 99, 235, 0.3);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>
            Talkkonnect XML Configuration
            {{if .Message}}
                <span class="status {{.StatusType}}">{{.Message}}</span>
            {{end}}
        </h1>
        <form action="/config" method="POST">
            <textarea name="xmlcontent" spellcheck="false">{{.Content}}</textarea>
            <div class="actions">
                <button type="submit">Save Configuration</button>
            </div>
        </form>
    </div>
</body>
</html>
`

type configPageData struct {
	Content    string
	Message    string
	StatusType string // "success" or "error"
}

func (b *Talkkonnect) httpConfig(w http.ResponseWriter, r *http.Request) {
	if !remoteControlHTTPClientIPAllowed(r) {
		log.Printf("error: HTTP config UI request from %q rejected by remote control network ACL\n", r.RemoteAddr)
		http.Error(w, "403 forbidden: client address not allowed by remote control network ACL", http.StatusForbidden)
		return
	}

	tmpl, err := template.New("config").Parse(configTemplate)
	if err != nil {
		http.Error(w, "Failed to parse template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := configPageData{}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			data.Message = "Failed to parse form: " + err.Error()
			data.StatusType = "error"
		} else {
			newXMLContent := r.FormValue("xmlcontent")

			// Save the new XML content back to the file
			err := ioutil.WriteFile(ConfigXMLFile, []byte(newXMLContent), 0644)
			if err != nil {
				data.Message = "Failed to write config file: " + err.Error()
				data.StatusType = "error"
			} else {
				log.Println("success: Received and updated Talkkonnect config from web interface.")

				// Attempt to live-reload the internal application state
				err = readxmlconfig(ConfigXMLFile, true)
				if err != nil {
					data.Message = "Config saved, but validation/reload failed: " + err.Error()
					data.StatusType = "error"
					log.Println("error: Validation/reload failed after saving web config:", err)
				} else {
					CheckConfigSanity(true)
					data.Message = "Configuration saved and reloaded successfully!"
					data.StatusType = "success"
				}
			}
		}
	}

	// Read current file context
	fileContent, err := ioutil.ReadFile(ConfigXMLFile)
	if err != nil {
		if data.Message == "" {
			data.Message = "Loaded default UI context for config string, but reading raw XML failed: " + err.Error()
			data.StatusType = "error"
		}
	} else {
		data.Content = string(fileContent)
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println("error: Failed to execute template:", err)
	}
}
