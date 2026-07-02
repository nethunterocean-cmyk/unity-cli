# unity-cli

[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [Français](README.fr.md)

> Contrôle Unity Editor depuis la ligne de commande. Fait pour les agents AI, mais ça marche avec n'importe quoi, tabarnak.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Pas de serveur à lancer. Pas de config à écrire. Pas de processus à gérer. Juste une commande, c'est tout.**

> **🔒 Fork de sécurité:** Toutes les fonctionnalités de mise à jour automatique et de vérification de version ont été retirées pour des raisons de sécurité.
> Pour la version originale avec mises à jour automatiques, c'est par là → [youngwoocho02/unity-cli](https://github.com/youngwoocho02/unity-cli).

## Build (CLI)

```bash
git clone https://github.com/nethunterocean-cmyk/unity-cli
cd unity-cli
go build -o unity-cli .
sudo mv unity-cli /usr/local/bin/
```

Requiert [Go](https://go.dev/dl/) 1.24+. Pas d'autres dépendances, ostie.

## Configuration Unity

Copie le dossier `UnityFiles/` dans `Assets/` de ton projet Unity:

```bash
cp -r UnityFiles /path/to/YourUnityProject/Assets/
```

Le connecteur démarre automatiquement quand Unity s'ouvre. Aucune configuration nécessaire. Faque là, tu peux gérer ton Unity en ligne de commande, crisse de belle affaire.
