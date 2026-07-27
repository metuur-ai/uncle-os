export interface InstallStep {
  stepNumber: number;
  title: string;
  command: string;
  description: string;
  keyRule: string;
  mockTerminalOutput: string;
}

export interface InstallArtifact {
  filename: string;
  runsOn: string;
  detectedAs: string;
}

export interface InstallOption {
  envVar: string;
  defaultVal: string;
  description: string;
  example: string;
}

export interface InstallPath {
  id: string;
  label: string;
  badge: string;
  summary: string;
  requirement: string;
  commands: string[];
}

export interface InstallFaq {
  id: string;
  question: string;
  answer: string;
  command?: string;
}

/** The one-line install. Everything else on this page is a variation on it. */
export const INSTALL_ONE_LINER =
  'curl -fsSL https://raw.githubusercontent.com/metuur-ai/uncle-os/main/company-os-starter/install.sh | bash';

export const INSTALL_STEPS: InstallStep[] = [
  {
    stepNumber: 1,
    title: 'Install the CLI',
    command: INSTALL_ONE_LINER,
    description:
      'Detects your OS and architecture, fetches the matching binary into ~/.local/bin, then runs --version to prove it works before reporting success.',
    keyRule:
      'A single static binary. No interpreter, no runtime, no package manager underneath it.',
    mockTerminalOutput: `Installing company-os
  CLI:    /Users/you/.local/bin/company-os
  Downloading https://github.com/metuur-ai/uncle-os/releases/latest/download/company-os-darwin-arm64
  installed company-os 1.4.0 (commit a1b2c3d, go1.25.7, darwin/arm64)

Next
  company-os --help                  # the whole surface
  cd <a workspace root>              # or pass --root everywhere
  company-os validate                # the CI gate`,
  },
  {
    stepNumber: 2,
    title: 'Put it on your PATH',
    command:
      'echo \'export PATH="$HOME/.local/bin:$PATH"\' >> ~/.zshrc && source ~/.zshrc',
    description:
      'The installer warns you if ~/.local/bin is missing from PATH and prints this line. Skip this step if it did not warn you — you are already set.',
    keyRule:
      'Nothing is installed system-wide. Relocate the binary with INSTALL_DIR instead of using sudo.',
    mockTerminalOutput: `  ~/.local/bin is not on your PATH — company-os won't be found until you add it.
  zsh:  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc  && source ~/.zshrc
  bash: echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc`,
  },
  {
    stepNumber: 3,
    title: 'Verify the install',
    command: 'company-os --version',
    description:
      'Prints the release version, the commit it was built from, the Go toolchain, and the platform. A binary is never ambiguous about what it is.',
    keyRule:
      'If this prints a usage banner and exits 2, an old Python launcher is shadowing the binary — see the FAQ below.',
    mockTerminalOutput: `company-os 1.4.0 (commit a1b2c3d, go1.25.7, darwin/arm64)`,
  },
  {
    stepNumber: 4,
    title: 'Scaffold your first workspace',
    command:
      'mkdir moonbeam-os && cd moonbeam-os\ncompany-os init --company "Moonbeam Bakery" --team web --platform ordering',
    description:
      'Creates the four peer roots every workspace is built on: company-os/, platforms/<p>/, teams/<t>/, and company-ontology/.',
    keyRule:
      'init refuses to run inside an existing workspace root. Passing all three flags skips the interactive prompts.',
    mockTerminalOutput: `initialized workspace at /Users/you/moonbeam-os
  company: Moonbeam Bakery | first team: web | first platform: ordering
next: cd /Users/you/moonbeam-os && company-os discover new --team web "<discovery title>"`,
  },
];

export const INSTALL_ARTIFACTS: InstallArtifact[] = [
  {
    filename: 'company-os-darwin-arm64',
    runsOn: 'Apple Silicon Macs',
    detectedAs: 'Darwin / arm64',
  },
  {
    filename: 'company-os-darwin-amd64',
    runsOn: 'Intel Macs',
    detectedAs: 'Darwin / x86_64',
  },
  {
    filename: 'company-os-linux-amd64',
    runsOn: 'x86-64 Linux',
    detectedAs: 'Linux / x86_64',
  },
];

export const INSTALL_OPTIONS: InstallOption[] = [
  {
    envVar: 'INSTALL_DIR',
    defaultVal: '~/.local/bin',
    description:
      'Where the binary lands. The installer elevates with sudo only if the directory is unwritable.',
    example: `INSTALL_DIR=/usr/local/bin bash -c "$(curl -fsSL .../install.sh)"`,
  },
  {
    envVar: 'VERSION',
    defaultVal: 'latest',
    description:
      'Pin a release tag instead of tracking latest. Useful for matching a teammate mid-federation.',
    example: 'VERSION=v0.4.0 bash -c "$(curl -fsSL .../install.sh)"',
  },
  {
    envVar: 'BASE_URL',
    defaultVal: 'github.com/metuur-ai/uncle-os/releases/latest/download',
    description:
      'Override the download base — for an internal mirror or an air-gapped artifact store.',
    example: 'BASE_URL=https://mirror.internal/company-os bash install.sh',
  },
];

export const INSTALL_PATHS: InstallPath[] = [
  {
    id: 'installer',
    label: 'Installer script',
    badge: 'RECOMMENDED',
    summary:
      'One line. Detects your platform, downloads the release artifact, verifies it runs.',
    requirement: 'curl or wget. Nothing else.',
    commands: [INSTALL_ONE_LINER],
  },
  {
    id: 'source',
    label: 'Build from source',
    badge: 'GO TOOLCHAIN',
    summary:
      'Compiles the binary and copies it to ~/.local/bin. Override the destination with PREFIX.',
    requirement:
      'The Go toolchain — a build-time requirement only. The binary it produces has no dependencies.',
    commands: [
      'git clone https://github.com/metuur-ai/uncle-os',
      'cd uncle-os/company-os-starter',
      'make install            # PREFIX=/some/where to relocate',
    ],
  },
  {
    id: 'checkout',
    label: 'From a checkout or unpacked release',
    badge: 'OFFLINE',
    summary:
      'Run the same script with no network: it prefers a binary already sitting in dist/ over downloading one.',
    requirement: 'A cloned repo or an unpacked release bundle.',
    commands: ['cd company-os-starter', 'bash install.sh'],
  },
];

export const INSTALL_FAQS: InstallFaq[] = [
  {
    id: 'macos-gatekeeper',
    question: 'macOS: why an installer instead of a download link?',
    answer:
      'The releases are deliberately unsigned and un-notarized, and with this install path they do not need to be. com.apple.quarantine is attached by the downloading application — Safari, Chrome, Mail — never by curl, wget or tar, and Gatekeeper only adjudicates files carrying it. A binary fetched by the installer runs immediately, with no prompt. Download the same bytes in a browser and you get the quarantined path, which does not fail cleanly: it hangs with no output at all. If you took that route, clear the attribute once.',
    command: 'xattr -d com.apple.quarantine ~/.local/bin/company-os',
  },
  {
    id: 'spctl',
    question: 'spctl says "rejected" — is the binary broken?',
    answer:
      'No, and that is expected. spctl answers "would Gatekeeper admit this?", which is a different question from "will this execute?" — and Gatekeeper is never consulted for a file with no quarantine attribute. Check the attribute, not spctl. It should print nothing, or only com.apple.provenance.',
    command: 'xattr ~/.local/bin/company-os',
  },
  {
    id: 'checksums',
    question: 'How do I verify what I downloaded?',
    answer:
      'SHA256SUMS is published next to the artifacts on every release. The checksums are reproducible and that is checked in CI: two clean clones of the same commit, at different paths, produce byte-identical artifacts.',
    command:
      'shasum -a 256 -c SHA256SUMS --ignore-missing    # macOS\nsha256sum -c SHA256SUMS --ignore-missing        # Linux',
  },
  {
    id: 'upgrade',
    question: 'How do I upgrade?',
    answer:
      'Re-run the installer, or make install from a checkout. Both write to a sibling path and rename over the target — the rename is atomic, so no half-written binary ever sits on your PATH, and it unlinks the old inode rather than truncating it (which is what an in-place cp cannot do on Linux while the old binary is running). There is no migration step and no state outside the workspace.',
    command: INSTALL_ONE_LINER,
  },
  {
    id: 'shadowed',
    question: '--version prints a usage banner and exits 2',
    answer:
      'A leftover launcher from the old Python install.sh is earlier on your PATH and is silently shadowing the Go binary. Every real subcommand still succeeds, so nothing looks wrong — which is why this needs finding rather than waiting for a symptom. List the resolution order, remove the launcher and its kit root, then re-verify.',
    command:
      'type -a company-os                     # every company-os on PATH, in order\nrm -f ~/.local/bin/company-os          # the generated launcher\nrm -rf ~/.local/share/company-os       # the Python kit root\ncompany-os --version',
  },
  {
    id: 'uninstall',
    question: 'How do I uninstall?',
    answer:
      'Delete the binary. There is nothing else on disk — no kit directory, no config outside your workspaces, no package manager entry. From a checkout, make uninstall does the same thing.',
    command: 'rm -f ~/.local/bin/company-os          # or: make uninstall',
  },
  {
    id: 'windows',
    question: 'Is there a Windows build?',
    answer:
      'No, deliberately. Go cross-compiles to Windows for free, but nothing here is tested there, so nothing is claimed. Use WSL2 with the linux-amd64 artifact.',
  },
];

export const COMPANION_TOOL = {
  name: 'local-search',
  repo: 'https://github.com/metuur-ai/local-search',
  summary:
    'Offline BM25 search over your workspace docs — discovery briefs, PRDs, reality docs, ADRs. Optional, separate binary, same unsigned-and-curl-fetched install approach.',
  steps: [
    {
      command:
        'curl -fsSL https://raw.githubusercontent.com/metuur-ai/local-search/main/install.sh | bash',
      description:
        'Installs the local-search CLI, its agent skill into ~/.claude/skills/local-search, and a web UI into ~/.local/share/local-search/web.',
    },
    {
      command: 'local-search repo add ~/moonbeam-os moonbeam-os',
      description: 'Register the workspace you just scaffolded so it gets indexed.',
    },
    {
      command: 'local-search install-skill --global',
      description:
        'Teach every agent session how to find existing workspace docs before writing new ones.',
    },
  ],
};
