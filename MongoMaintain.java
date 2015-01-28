import com.google.common.io.*;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class MongoMaintain {
    public static final String MONGO_MAINTAIN = "mongo-maintain";
    private static final String DOWNLOAD_URL_PREFIX = "http://4sh-projects.github.io/mongo-maintain/versions/latest/downloads/";
    private final boolean isWindows;
    private final boolean isMac;
    private final boolean isLinux;
    private String architecture;
    private String mongoMaintainPath;

    public MongoMaintain() {
        String osName = System.getProperty("os.name");
        isWindows = osName.startsWith("Windows");
        isMac = osName.startsWith("Mac OS X") || osName.startsWith("Darwin");
        isLinux = (osName.contains("nix") || osName.contains("nux") || osName.indexOf("aix") > 0 );

        architecture = System.getProperty("os.arch");

        // See http://stackoverflow.com/questions/4748673/how-can-i-check-the-bitness-of-my-os-using-java-j2se-not-os-arch
        if (isWindows) {
            String arch = System.getenv("PROCESSOR_ARCHITECTURE");
            String wow64Arch = System.getenv("PROCESSOR_ARCHITEW6432");

            architecture = arch.endsWith("64")
                    || wow64Arch != null && wow64Arch.endsWith("64")
                    ? "amd64" : "x86";
        } else {
            architecture = architecture.endsWith("64") ? "amd64" : "x86";
        }
    }

    /**
     * Will download mongo-maintain if not already installed. If download is needed, then the latest version will
     * be installed. Then mongo-maintain will be run with your args.
     *
     * Mac users : you maybe be will have execution error because of mongodump. If mongodump is not installed in
     * /usr/local/bin then you can specify mongoHome vm property (ex: -DmongoHome=/Users/dro/dev/tools/mongo/bin).
     */
    public void run(MongoMaintainParams params) throws IOException, InterruptedException {
        if (mongoMaintainPath == null) {
            mongoMaintainPath = download().getAbsolutePath();
        }

        List<String> args = new ArrayList<>();

        args.add(mongoMaintainPath);

        if (params.getScriptsFolder() != null) {
            args.add("-scripts=" + params.getScriptsFolder());
        }

        if (params.getUrl() != null) {
            args.add("-url=" + params.getUrl());
        }

        if (params.getDatabase() != null) {
            args.add("-database=" + params.getDatabase());
        }

        if (params.getUser() != null) {
            args.add("-user=" + params.getUser());
        }

        if (params.getPassword() != null) {
            args.add("-password=" + params.getPassword());
        }

        System.out.println("Execute mongo-maintain with cmd :");
        System.out.println(args);

        ProcessBuilder processBuilder = new ProcessBuilder(args);
        int exitValue = processBuilder
                    .redirectOutput(ProcessBuilder.Redirect.INHERIT)
                    .redirectError(ProcessBuilder.Redirect.INHERIT)
                    .start()
                    .waitFor();

        if (exitValue != 0) {
            String mongoHome = "";

            if (isMac) {
                mongoHome = System.getProperty("mongoHome");

                if (mongoHome == null || mongoHome.isEmpty()) {
                    mongoHome = "/usr/local/bin";
                }

                System.out.println("Trying to re-execute mongo-maintain with a new $PATH containing mongoHome " + mongoHome);

                processBuilder = new ProcessBuilder(args);

                Map<String, String> env = processBuilder.environment();
                env.put("PATH", env.get("PATH") + ":" + mongoHome);

                exitValue = processBuilder
                        .redirectOutput(ProcessBuilder.Redirect.INHERIT)
                        .redirectError(ProcessBuilder.Redirect.INHERIT)
                        .start()
                        .waitFor();
            }

            if (exitValue != 0) {
                if (isMac) {
                    if ("/usr/local/bin".equals(mongoHome)) {
                        System.err.print("If you have error because of mongodump, " +
                                "you can add a specific specific path to access mongodump by setting property " +
                                "mongoHome as vm arg (ex: -DmongoHome=/Users/dro/dev/tools/mongo/bin)");
                    } else {
                        System.err.print("If you have error because of mongodump, we have no more solution. " +
                                "Check again your -DmongoHome or try to create a symbolic link of mongodump " +
                                "in /usr/local/bin (and remove -DmongoHome)");
                    }
                }

                throw new AssertionError("");
            }
        }
    }

    private File download() throws IOException {
        File installDir = new File(new File("/tmp"), "mongo-maintain");

        String url;
        File mongoMaintainExe;

        if (isWindows) {
            url = DOWNLOAD_URL_PREFIX + "windows-" + architecture + "/" + MONGO_MAINTAIN + ".exe";
            mongoMaintainExe = new File(installDir, MONGO_MAINTAIN + ".exe");
        } else if (isMac) {
            url = DOWNLOAD_URL_PREFIX + "darwin-" + architecture + "/" + MONGO_MAINTAIN;
            mongoMaintainExe = new File(installDir, MONGO_MAINTAIN);
        } else if (isLinux) {
            url = DOWNLOAD_URL_PREFIX + "linux-" + architecture + "/" + MONGO_MAINTAIN;
            mongoMaintainExe = new File(installDir, MONGO_MAINTAIN);
        } else {
            throw new IllegalStateException(String.format("This OS not supported by mongo-maintain"));
        }

        if (mongoMaintainExe.exists()) {
            return mongoMaintainExe;
        }

        System.out.println("Downloading mongo-maintain from " + url + "...");

        try {
            System.out.println("mongo-maintain will be here : " + mongoMaintainExe.getAbsolutePath());
            mongoMaintainExe.getParentFile().mkdirs();

            InputSupplier<InputStream> input = Resources.newInputStreamSupplier(URI.create(url).toURL());
            OutputSupplier<FileOutputStream> ouput = Files.newOutputStreamSupplier(mongoMaintainExe);

            ByteStreams.copy(input, ouput);
        } catch (IOException e) {
            throw new IllegalStateException("Unable to mongo-maintain from " + url);
        }

        if (isLinux) {
            Runtime.getRuntime().exec("chmod +x " + mongoMaintainExe.getAbsolutePath());
        } else {
            mongoMaintainExe.setExecutable(true);
        }

        return mongoMaintainExe;
    }

    public static class MongoMaintainParams {
        /**
         * The absolute path of your folder containing scripts
         */
        private String scriptsFolder;

        /**
         * The url to connect to mongo. Example : localhost:27017
         */
        private String url;

        /**
         * The database to connect to.
         */
        private String database;

        /**
         * The username to use to connect to mongo. Optional
         */
        private String user;

        /**
         * The password to use to connect to mongo. Optional, but shoulb be set of user is set
         */
        private String password;

        public MongoMaintainParams(String scriptsFolder, String url, String database) {
            this.scriptsFolder = scriptsFolder;
            this.url = url;
            this.database = database;
            this.user = null;
            this.password = null;
        }

        public MongoMaintainParams(String scriptsFolder, String url, String database, String user, String password) {
            this.scriptsFolder = scriptsFolder;
            this.url = url;
            this.database = database;
            this.user = user;
            this.password = password;
        }

        public String getScriptsFolder() {
            return scriptsFolder;
        }

        public String getUrl() {
            return url;
        }

        public String getDatabase() {
            return database;
        }

        public String getUser() {
            return user;
        }

        public String getPassword() {
            return password;
        }
    }
}
