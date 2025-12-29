import pytest
from lazyworktree.security import TrustManager, TrustStatus


@pytest.fixture
def trust_db_path(tmp_path):
    return tmp_path / "trusted.json"


@pytest.fixture
def trust_manager(trust_db_path):
    return TrustManager(db_path=trust_db_path)


def test_check_trust_not_found(trust_manager, tmp_path):
    non_existent = tmp_path / "ghost.wt"
    assert trust_manager.check_trust(non_existent) == TrustStatus.NOT_FOUND


def test_trust_flow(trust_manager, tmp_path):
    wt_file = tmp_path / "test.wt"
    wt_file.write_text("echo 'hello'", encoding="utf-8")

    # Initially untrusted
    assert trust_manager.check_trust(wt_file) == TrustStatus.UNTRUSTED

    # Trust it
    trust_manager.trust_file(wt_file)
    assert trust_manager.check_trust(wt_file) == TrustStatus.TRUSTED

    # Persisted?
    new_manager = TrustManager(db_path=trust_manager.db_path)
    assert new_manager.check_trust(wt_file) == TrustStatus.TRUSTED


def test_modification_revokes_trust(trust_manager, tmp_path):
    wt_file = tmp_path / "test.wt"
    wt_file.write_text("echo 'initial'", encoding="utf-8")
    trust_manager.trust_file(wt_file)

    assert trust_manager.check_trust(wt_file) == TrustStatus.TRUSTED

    # Modify file
    wt_file.write_text("echo 'modified'", encoding="utf-8")
    assert trust_manager.check_trust(wt_file) == TrustStatus.UNTRUSTED


def test_corrupt_db_resets(trust_db_path, tmp_path):
    # Create corrupt JSON
    trust_db_path.write_text("{invalid_json", encoding="utf-8")

    manager = TrustManager(db_path=trust_db_path)
    wt_file = tmp_path / "test.wt"
    wt_file.write_text("echo 'safe'", encoding="utf-8")

    # Should default to empty/untrusted, not crash
    assert manager.check_trust(wt_file) == TrustStatus.UNTRUSTED
