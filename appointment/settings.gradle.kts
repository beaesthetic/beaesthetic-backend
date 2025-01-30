pluginManagement {
  val quarkusPluginVersion: String by settings
  val quarkusPluginId: String by settings
  repositories {
    gradlePluginPortal()
    mavenCentral()
    mavenLocal()
  }
  plugins { id(quarkusPluginId) version quarkusPluginVersion }
}
