import {
  Home,
  Download,
  FolderTree,
  Terminal,
  PlayCircle,
  ShieldCheck,
  CheckSquare,
  Search,
  BookOpen,
  Award,
  type LucideIcon,
} from 'lucide-react';
import { TabType } from './types';

/**
 * Single source of truth for the tab sequence.
 *
 * Navbar, HomeOverview, the content directory and every LessonFooter read from
 * this list, so the ordering, labels and "why" copy cannot drift apart. Copy is
 * carried over verbatim from the original Navbar definitions.
 */
export interface Lesson {
  id: TabType;
  label: string;
  /** Compact label for the sticky nav rail. */
  shortLabel: string;
  icon: LucideIcon;
  whyText: string;
}

export const LESSONS: Lesson[] = [
  {
    id: 'home',
    label: 'Index & Overview',
    shortLabel: 'Overview',
    icon: Home,
    whyText:
      'Start here for a high-level summary of how Company OS connects product, engineering, and governance.',
  },
  {
    id: 'install',
    label: 'Install & Setup',
    shortLabel: 'Install',
    icon: Download,
    whyText:
      'Install the company-os CLI in one line, put it on your PATH, and scaffold your first workspace.',
  },
  {
    id: 'architecture',
    label: 'Workspace Architecture',
    shortLabel: 'Architecture',
    icon: FolderTree,
    whyText:
      'Explore the workspace directory structure, file hierarchy, and folder layouts for Company OS and Team OS.',
  },
  {
    id: 'cli',
    label: 'CLI Terminal Explorer',
    shortLabel: 'CLI',
    icon: Terminal,
    whyText:
      'Run interactive terminal commands like validate, graph build, and prd new in a live CLI simulator.',
  },
  {
    id: 'workflows',
    label: 'Interactive Workflows',
    shortLabel: 'Workflows',
    icon: PlayCircle,
    whyText:
      'Simulate a complete PRD lifecycle step-by-step from discovery brief to production release.',
  },
  {
    id: 'governance',
    label: 'Governance Tiers',
    shortLabel: 'Governance',
    icon: ShieldCheck,
    whyText:
      'Learn how Canonical, Team, and Personal rules coordinate without confusion or broken builds.',
  },
  {
    id: 'validation',
    label: 'Validation Gates (1-8)',
    shortLabel: 'Validation',
    icon: CheckSquare,
    whyText:
      'Inspect automated compliance gates (1-8) that verify your PRDs, dependencies, and workspace rules.',
  },
  {
    id: 'search-agent',
    label: 'Local Search & BM25',
    shortLabel: 'Search',
    icon: Search,
    whyText:
      'Search offline across workspace docs and discover AI Agent Skills for authoring standardized artifacts.',
  },
  {
    id: 'reference',
    label: 'Reference Matrix',
    shortLabel: 'Reference',
    icon: BookOpen,
    whyText:
      'Compare YAML configurations, schema rules, and precedence resolution matrices side-by-side.',
  },
  {
    id: 'quiz',
    label: 'Knowledge Check',
    shortLabel: 'Quiz',
    icon: Award,
    whyText: 'Test your knowledge with a scored quiz covering every core concept in the platform.',
  },
];

/** Lessons excluding home — the nine teaching tabs, in order. */
export const TUTORIALS = LESSONS.filter((l) => l.id !== 'home');

export const getLesson = (id: TabType): Lesson =>
  LESSONS.find((l) => l.id === id) ?? LESSONS[0];

/** 1-based position among the nine tutorials. Returns 0 for home. */
export const lessonNumber = (id: TabType): number =>
  TUTORIALS.findIndex((l) => l.id === id) + 1;

/** The lesson that follows `id`, wrapping from the last tutorial back to home. */
export const nextLesson = (id: TabType): Lesson => {
  const i = LESSONS.findIndex((l) => l.id === id);
  return LESSONS[(i + 1) % LESSONS.length];
};
