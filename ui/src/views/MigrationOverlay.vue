<template>
  <div class="modal is-active" v-if="isMigrating">
    <div class="modal-background"></div>
    <div class="modal-card">
      <section class="modal-card-body has-text-centered">
        <div class="content">
          <h3 class="title is-4">Database Maintenance in Progress</h3>
          <p class="subtitle is-6">Please wait while the database is being updated...</p>

          <div v-if="migrationState.current">
            <p class="migration-current">{{ migrationState.current }}</p>
          </div>

          <div v-if="migrationState.message" class="migration-message">
            <p>{{ migrationState.message }}</p>
          </div>

          <div v-if="migrationState.total > 0" class="migration-progress">
            <progress
              class="progress is-primary"
              :value="migrationState.progress"
              :max="migrationState.total"
            >
              {{ progressPercent }}%
            </progress>
            <p class="progress-text">
              {{ migrationState.progress }} / {{ migrationState.total }}
            </p>
          </div>

          <div v-else>
            <progress class="progress is-primary" max="100"></progress>
          </div>

          <p class="help-notice mt-4">
            Web interface will become available when operation is complete.
          </p>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import ky from 'ky'

export default {
  name: 'MigrationOverlay',
  data() {
    return {
      isMigrating: false,
      migrationState: {
        is_running: false,
        current: '',
        total: 0,
        progress: 0,
        message: ''
      },
      pollInterval: null
    }
  },
  computed: {
    progressPercent() {
      if (this.migrationState.total === 0) return 0
      return Math.round((this.migrationState.progress / this.migrationState.total) * 100)
    }
  },
  async mounted() {
    await this.checkMigrationStatus()
    // Only start polling if migrations are actually running
    if (this.isMigrating) {
      this.startPolling()
    }
  },
  beforeDestroy() {
    this.stopPolling()
  },
  methods: {
    async checkMigrationStatus() {
      try {
        const response = await ky.get('/api/options/state').json()
        if (response.currentState && response.currentState.migration) {
          this.migrationState = response.currentState.migration
          this.isMigrating = response.currentState.migration.is_running

          // Stop polling if migrations are complete
          if (!response.currentState.migration.is_running) {
            this.stopPolling()
          }
        }
      } catch (error) {
        console.error('Failed to check migration status:', error)
      }
    },
    startPolling() {
      this.pollInterval = setInterval(() => {
        this.checkMigrationStatus()
      }, 2000)
    },
    stopPolling() {
      if (this.pollInterval) {
        clearInterval(this.pollInterval)
        this.pollInterval = null
      }
    }
  }
}
</script>

<style scoped>
.modal-card {
  max-width: 600px;
}

.migration-current {
  font-weight: 600;
  margin-bottom: 1rem;
  color: #363636;
}

.migration-message {
  margin: 1rem 0;
  padding: 0.75rem;
  background-color: #f5f5f5;
  border-radius: 4px;
  font-size: 0.9rem;
}

.migration-progress {
  margin-top: 1.5rem;
}

.progress-text {
  margin-top: 0.5rem;
  font-size: 0.9rem;
  color: #7a7a7a;
}

.help-notice {
  font-size: 1.1rem;
  font-weight: 500;
  color: #3273dc;
}
</style>
