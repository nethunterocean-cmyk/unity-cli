# unity-cli

[English](README.md) | [한국어](README.ko.md) | [日本語](README.ja.md) | [Français](README.fr.md)

> Contrôle Unity Editor depuis la ligne de commande. Fait pour les agents AI, mais ça marche avec n'importe quoi.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Pas de serveur à lancer. Pas de config à écrire. Pas de processus à gérer. Juste une commande.**

> **🔒 Fork de sécurité:** Toutes les fonctionnalités de mise à jour automatique et de vérification de version ont été retirées pour des raisons de sécurité.
> Pour la version originale avec mises à jour automatiques → [youngwoocho02/unity-cli](https://github.com/youngwoocho02/unity-cli).

## Build (CLI)

```bash
git clone https://github.com/nethunterocean-cmyk/unity-cli
cd unity-cli
go build -o unity-cli .
sudo mv unity-cli /usr/local/bin/
```

Requiert [Go](https://go.dev/dl/) 1.24+. Pas d'autres dépendances.

## Configuration Unity

Copie le dossier `UnityFiles/` dans le répertoire `Assets/` de ton projet Unity:

```bash
cp -r UnityFiles /path/to/YourUnityProject/Assets/
```

Le connecteur démarre automatiquement quand Unity s'ouvre. Aucune configuration nécessaire.

### Recommandé: Désactiver le Throttling de l'Éditeur

Par défaut, Unity réduit les mises à jour de l'éditeur quand la fenêtre n'est pas au premier plan. Cela peut retarder les commandes CLI car le travail API Unity est dispatché sur le thread principal de l'éditeur.

Pour corriger ça, va dans **Edit → Preferences → General → Interaction Mode** et mets-le à **No Throttling**.

Le connecteur demande aussi une mise à jour PlayerLoop à chaque requête CLI. No Throttling est tout de même recommandé pour le comportement le plus réactif en arrière-plan.

## Démarrage Rapide

```bash
# Vérifier la connexion Unity
unity-cli status

# Entrer en mode play et attendre
unity-cli editor play --wait

# Exécuter du code C# dans Unity
unity-cli exec "return Application.dataPath;"

# Lire les logs console
unity-cli console --type error,warning,log
```

## Comment ça marche

```
Terminal                              Unity Editor
────────                              ────────────
$ unity-cli editor play --wait
    │
    ├─ scanne ~/.unity-cli/instances/*.json
    │  → sélectionne l'instance Unity pour ce projet
    │
    ├─ envoie la commande au listener Unity sélectionné
    │  { "command": "manage_editor",
    │    "params": { "action": "play",
    │                "wait_for_completion": true }}
    │                                      │
    │                                  HttpServer reçoit
    │                                      │
    │                                  CommandRouter dispatch
    │                                      │
    │                                  ManageEditor.HandleCommand()
    │                                  → EditorApplication.isPlaying = true
    │                                  → attend PlayModeStateChange
    │                                      │
    ├─ reçoit la réponse JSON  ←───────────┘
    │  { "success": true,
    │    "message": "Entered play mode (confirmed)." }
    │
    └─ affiche: Entered play mode (confirmed).
```

Le Connecteur Unity:
1. Ovre un listener HTTP local au démarrage de l'Éditeur
2. Écrit un fichier d'instance par projet dans `~/.unity-cli/instances/` pour que la CLI sache où se connecter
3. Met à jour le fichier d'instance toutes les 0.5s avec l'état courant (heartbeat)
4. Découvre toutes les classes `[UnityCliTool]` via la réflexion à chaque requête
5. Route les commandes entrantes vers le handler correspondant sur le thread principal
6. Survit aux rechargements de domaine (recompilation de scripts)

Avant de compiler ou recharger, le Connecteur enregistre l'état (`compiling`, `reloading`) dans le fichier d'instance. Quand le thread principal se fige, le timestamp arrête de bouger. La CLI détecte ça et attend un timestamp frais avant d'envoyer des commandes.

## Commandes Intégrées

| Commande | Description |
|----------|-------------|
| `editor` | Contrôler play/stop/pause/refresh de Unity Editor |
| `console` | Lire, filtrer et effacer les logs console |
| `exec` | Exécuter du code C# arbitraire dans Unity |
| `test` | Lancer les tests EditMode/PlayMode |
| `menu` | Exécuter un élément de menu Unity par chemin |
| `reserialize` | Re-sérialiser les assets via le sérialiseur Unity |
| `screenshot` | Capturer la vue Scene/Game en PNG |
| `profiler` | Lire la hiérarchie du profiler, contrôler l'enregistrement |
| `list` | Afficher tous les outils disponibles avec leurs schémas de paramètres |
| `status` | Afficher l'état de connexion Unity Editor |
| `update` | Mise à jour automatique — désactivée dans ce fork pour sécurité |

### Contrôle de l'Éditeur

```bash
# Entrer en mode play
unity-cli editor play

# Entrer en mode play et attendre le chargement complet
unity-cli editor play --wait

# Sortir du mode play
unity-cli editor stop

# Activer/désactiver la pause (mode play uniquement)
unity-cli editor pause

# Rafraîchir les assets (bloqué en mode play sans --force)
unity-cli editor refresh

# Rafraîchir et recompiler les scripts
unity-cli editor refresh --compile

# Forcer le rafraîchissement pendant le mode play
unity-cli editor refresh --force
```

### Logs Console

```bash
# Lire les logs d'erreur et d'avertissement (défaut)
unity-cli console

# Lire les 20 dernières entrées de tous types
unity-cli console --lines 20 --filter error,warning,log

# Lire seulement les erreurs
unity-cli console --type error

# Avec stack trace (user: code utilisateur seulement, full: brut)
unity-cli console --stacktrace user

# Effacer la console
unity-cli console --clear
```

### Exécuter du Code C#

Exécute du code C# arbitraire dans Unity Editor au runtime. C'est la commande la plus puissante — elle donne accès complet à UnityEngine, UnityEditor, ECS et tous les assemblies chargés. Pas besoin d'écrire un outil personnalisé pour des requêtes ou mutations ponctuelles.

Utilise `return` pour obtenir une sortie. Les espaces de noms courants sont inclus par défaut. Ajoute `--usings` seulement pour les types spécifiques au projet (ex: `Unity.Entities`). `--usings` accepte des espaces de noms séparés par des virgules et peut être répété.

```bash
unity-cli exec "return Application.dataPath;"
unity-cli exec "return EditorSceneManager.GetActiveScene().name;"
unity-cli exec "return World.All.Count;" --usings Unity.Entities
unity-cli exec "return World.All.Count;" --usings Unity.Entities --usings Unity.Mathematics

# Pipe via stdin pour éviter les problèmes d'échappement shell
echo 'Debug.Log("hello"); return null;' | unity-cli exec
echo 'var go = new GameObject("Marker"); go.tag = "EditorOnly"; return go.name;' | unity-cli exec
```

`exec` bloque par défaut le code asynchrone, les coroutines et les callbacks différés d'Unity parce que la commande retourne avant que ces chemins se terminent. Utilise `--allow-async` seulement quand ce comportement différé est intentionnel.

Parce que `exec` compile et exécute du vrai C#, il peut faire tout ce qu'un outil personnalisé peut faire — inspecter des entités ECS, modifier des assets, appeler des API internes, exécuter des utilitaires d'édition. Pour les agents AI, ça signifie **un accès zéro-friction à tout le runtime Unity** sans écrire une seule ligne de code d'outil. Le pipe via stdin évite les maux de tête d'échappement shell avec du code complexe.

### Éléments de Menu

```bash
# Exécuter un élément de menu Unity par son chemin
unity-cli menu "File/Save Project"
unity-cli menu "Assets/Refresh"
unity-cli menu "Window/General/Console"
```

Note: `File/Quit` est bloqué pour des raisons de sécurité.

### Reserialize d'Assets

Les agents AI (et les humains) peuvent éditer les fichiers d'asset Unity — `.prefab`, `.unity`, `.asset`, `.mat` — en YAML texte brut. Mais le sérialiseur YAML d'Unity est strict: un champ manquant, une indentation erronée ou un `fileID` périmé peut corrompre l'asset silencieusement.

`reserialize` résout ça. Après une édition texte, il dit à Unity de charger l'asset en mémoire et de le réécrire via son propre sérialiseur. Le résultat est un fichier YAML propre et valide — comme si tu l'avais édité via l'Inspector.

```bash
# Reserialiser tout le projet (sans arguments)
unity-cli reserialize

# Après avoir édité les valeurs Transform d'un prefab dans un éditeur texte
unity-cli reserialize Assets/Prefabs/Player.prefab

# Après avoir modifié plusieurs scènes en lot
unity-cli reserialize Assets/Scenes/Main.unity Assets/Scenes/Lobby.unity

# Après avoir modifié des propriétés de matériau
unity-cli reserialize Assets/Materials/Character.mat
```

C'est ce qui rend l'édition d'assets basée sur le texte sûre. Sans ça, un seul champ YAML mal placé peut casser un prefab sans erreur visible jusqu'au runtime. Avec ça, **les agents AI peuvent modifier en toute confiance n'importe quel asset Unity via le texte brut** — ajouter des composants à des prefabs, ajuster des hiérarchies de scène, changer des propriétés de matériau — et savoir que le résultat se chargera correctement.

### Profiler

```bash
# Lire la hiérarchie du profiler (dernière frame, niveau supérieur)
unity-cli profiler hierarchy

# Drill-down récursif
unity-cli profiler hierarchy --depth 3

# Définir la racine par nom (correspondance partielle) — focus sur un système spécifique
unity-cli profiler hierarchy --root SimulationSystem --depth 3

# Drill-down dans un élément spécifique par ID
unity-cli profiler hierarchy --parent 4 --depth 2

# Moyenne sur les 30 dernières frames
unity-cli profiler hierarchy --frames 30 --min 0.5

# Moyenne sur une plage de frames spécifique
unity-cli profiler hierarchy --from 100 --to 200

# Filtrer et trier
unity-cli profiler hierarchy --min 0.5 --sort self --max 10

# Activer/désactiver l'enregistrement du profiler
unity-cli profiler enable
unity-cli profiler disable

# Afficher l'état du profiler
unity-cli profiler status

# Effacer les frames capturées
unity-cli profiler clear
```

### Lancer des Tests

Exécute les tests EditMode et PlayMode via le Unity Test Framework.

```bash
# Lancer les tests EditMode (défaut)
unity-cli test

# Lancer les tests PlayMode
unity-cli test --mode PlayMode

# Filtrer par nom de test (correspondance partielle)
unity-cli test --filter MyTestClass
```

Nécessite le package Unity Test Framework. Les tests PlayMode déclenchent un rechargement de domaine; la CLI interroge les résultats automatiquement.

### Liste des Outils

```bash
# Afficher tous les outils disponibles (intégrés + projet personnalisé) avec leurs schémas de paramètres
unity-cli list
```

### Outils Personnalisés

```bash
# Appeler un outil personnalisé directement par son nom
unity-cli my_custom_tool

# Appeler avec des paramètres
unity-cli my_custom_tool --params '{"key": "value"}'
```

### Statut

```bash
# Afficher l'état de Unity Editor
unity-cli status
# Sortie: Unity: ready
#   Project: /path/to/project
#   Version: 6000.1.0f1
#   PID:     12345
```

La CLI vérifie aussi automatiquement l'état d'Unity avant d'envoyer une commande. Si Unity est occupé (compilation, rechargement), elle attend qu'Unity devienne réactif.

## Options Globales

| Flag | Description | Défaut |
|------|-------------|--------|
| `--project <path>` | Sélectionner l'instance Unity par chemin de projet | auto |
| `--timeout <ms>` | Timeout de requête HTTP | 120000 |
| `--ignore-version-mismatch` | Ignorer la vérification de version CLI/connecteur | false |

```bash
# Sélectionner par chemin de projet quand plusieurs instances Unity sont ouvertes
unity-cli --project MyGame editor stop

# Exécuter même si les versions CLI et connecteur diffèrent
unity-cli --ignore-version-mismatch status
```

Utilise `--help` sur n'importe quelle commande pour une utilisation détaillée:

```bash
unity-cli editor --help
unity-cli exec --help
unity-cli profiler --help
```

## Créer des Outils Personnalisés

Crée une classe statique avec l'attribut `[UnityCliTool]` dans n'importe quel assembly Editor. Le Connecteur la découvre automatiquement au rechargement de domaine.

```csharp
using UnityCliConnector;
using Newtonsoft.Json.Linq;
using UnityEngine;

[UnityCliTool(Name = "spawn", Description = "Faire apparaître un ennemi à une position", Group = "gameplay")]
public static class SpawnEnemy
{
    public class Parameters
    {
        [ToolParameter("Position X dans le monde", Required = true)]
        public float X { get; set; }

        [ToolParameter("Position Y dans le monde", Required = true)]
        public float Y { get; set; }

        [ToolParameter("Position Z dans le monde", Required = true)]
        public float Z { get; set; }

        [ToolParameter("Nom du prefab dans le dossier Resources", DefaultValue = "Enemy")]
        public string Prefab { get; set; }
    }

    public static object HandleCommand(JObject parameters)
    {
        var p = new ToolParams(parameters);
        float x = p.GetFloat("x", 0);
        float y = p.GetFloat("y", 0);
        float z = p.GetFloat("z", 0);
        string prefabName = p.Get("prefab", "Enemy");

        var prefab = Resources.Load<GameObject>(prefabName);
        var instance = Object.Instantiate(prefab, new Vector3(x, y, z), Quaternion.identity);

        return new SuccessResponse("Enemy spawned", new
        {
            name = instance.name,
            position = new { x, y, z }
        });
    }
}
```

Appelle-le directement avec des flags ou du JSON:

```bash
unity-cli spawn --x 1 --y 0 --z 5 --prefab Goblin
unity-cli spawn --params '{"x":1,"y":0,"z":5,"prefab":"Goblin"}'
```

**Points clés:**

- **Nom**: sans `Name`, dérivé automatiquement du nom de classe (`SpawnEnemy` → `spawn_enemy`, `UITree` → `ui_tree`). Avec `Name = "spawn"`, la commande devient `unity-cli spawn`.
- **Classe Parameters**: optionnelle mais recommandée. `unity-cli list` l'utilise pour exposer les noms, types, descriptions et flags requis des paramètres — pour que les assistants AI puissent découvrir ton outil sans lire le source.
- **ToolParams**: utilise `p.Get()`, `p.GetInt()`, `p.GetFloat()`, `p.GetBool()`, `p.GetRaw()` pour une lecture cohérente des paramètres.
- **Découverte**: `unity-cli list` montre d'abord les outils intégrés (`group: "built-in"`), puis les outils personnalisés (`group: "custom"`) détectés depuis le projet Unity connecté.

**Référence des attributs:**

| Attribut | Propriété | Description |
|----------|-----------|-------------|
| `[UnityCliTool]` | `Name` | Surcharge du nom de commande (défaut: nom de classe → snake_case) |
| | `Description` | Description de l'outil affichée dans `list` |
| | `Group` | Nom du groupe pour la catégorisation |
| `[ToolParameter]` | `Description` | Description du paramètre (argument du constructeur) |
| | `Required` | Si le paramètre est requis (défaut: `false`) |
| | `Name` | Surcharge du nom du paramètre |
| | `DefaultValue` | Indication de valeur par défaut |

### Règles

- La classe doit être `static`
- Doit avoir `public static object HandleCommand(JObject parameters)` ou une variante `async Task<object>`
- Retourne `SuccessResponse(message, data)` ou `ErrorResponse(message)`
- Ajoute une classe `Parameters` imbriquée avec des attributs `[ToolParameter]` pour la découvrabilité
- Le nom de classe est automatiquement converti en snake_case pour le nom de commande
- Surcharge avec `[UnityCliTool(Name = "my_name")]` si nécessaire
- S'exécute sur le thread principal Unity, donc toutes les API Unity peuvent être appelées en toute sécurité
- Découvert automatiquement au démarrage de l'Éditeur et après chaque recompilation de script
- Les noms d'outils en double sont détectés et journalisés comme erreurs — seul le premier handler découvert est utilisé

## Instances Unity Multiples

Quand plusieurs Éditeurs Unity sont ouverts, chacun enregistre son chemin de projet:

```bash
# Voir toutes les instances en cours d'exécution
unity-cli status

# Sélectionner par chemin de projet
unity-cli --project MyGame editor play

# Défaut: utilise le projet Unity du répertoire de travail courant, ou la seule instance active
unity-cli editor play
```

## Comparé à MCP

| | MCP | unity-cli |
|---|-----|-----------|
| **Installation** | Python + uv + FastMCP + JSON de config | Binaire unique |
| **Dépendances** | Runtime Python, relais WebSocket | Aucune |
| **Protocole** | JSON-RPC 2.0 sur stdio + WebSocket | HTTP POST direct |
| **Configuration** | Générer config MCP, redémarrer outil AI | Ajouter le package Unity, terminé |
| **Reconnexion** | Logique complexe de reconnexion pour rechargements de domaine | Sans état par requête |
| **Support client** | Configuration client MCP uniquement | N'importe quoi avec un shell |
| **Outils personnalisés** | Même motif `[Attribute]` + `HandleCommand` | Identique |

## Auteur

Created by **DevBookOfArray**

[![YouTube](https://img.shields.io/badge/YouTube-DevBookOfArray-red?logo=youtube&logoColor=white)](https://www.youtube.com/@DevBookOfArray)
[![GitHub](https://img.shields.io/badge/GitHub-youngwoocho02-181717?logo=github)](https://github.com/youngwoocho02) (original)
[![GitHub](https://img.shields.io/badge/GitHub-nethunterocean--cmyk-181717?logo=github)](https://github.com/nethunterocean-cmyk/unity-cli) (security fork)

## Licence

MIT
